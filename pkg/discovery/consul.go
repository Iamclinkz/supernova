package discovery

import (
	"context"
	"fmt"
	"net/http"
	"supernova/pkg/conf"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

// 使用者consul配置使用
const (
	ConsulMiddlewareConfigConsulHostFieldName = "ConsulHost"
	ConsulMiddlewareConfigConsulPortFieldName = "ConsulPort"

	//ConsulRegisterConfigHealthcheckPortFieldName 指定consul心跳检查自己健康的Port
	ConsulRegisterConfigHealthcheckPortFieldName = "HealthCheckPort"
	ConsulRegisterConfigPushFieldName            = "Push"
)

// consul自用
const (
	consulMetaDataServiceProtocFieldName      = "X-Protoc-Type"
	consulMetaDataServiceExtraConfigFieldName = "X-Extra-Config"
)

type ConsulDiscoveryClient struct {
	Host             string
	Port             string
	client           consul.Client
	config           *api.Config
	mutex            sync.Mutex
	instancesMap     sync.Map //map[ServiceName][]*ServiceInstance
	httpServer       *http.Server
	middlewareConfig MiddlewareConfig
	registerConfig   RegisterConfig
	pushStopCh       chan struct{}
}

func NewConsulMiddlewareConfig(consulHost, consulPort string) MiddlewareConfig {
	return MiddlewareConfig{
		ConsulMiddlewareConfigConsulHostFieldName: consulHost,
		ConsulMiddlewareConfigConsulPortFieldName: consulPort,
	}
}

func NewConsulRegisterConfig(healthCheckPort string, push bool) RegisterConfig {
	ret := RegisterConfig{
		ConsulRegisterConfigHealthcheckPortFieldName: healthCheckPort,
	}
	if push {
		ret[ConsulRegisterConfigPushFieldName] = "1"
	}

	return ret
}

func newConsulDiscoveryClient(middlewareConfig MiddlewareConfig, registerConfig RegisterConfig) (DiscoverClient, error) {
	if middlewareConfig[ConsulMiddlewareConfigConsulHostFieldName] == "" ||
		middlewareConfig[ConsulMiddlewareConfigConsulPortFieldName] == "" {
		panic("")
	}
	consulConfig := api.DefaultConfig()
	consulConfig.Address = middlewareConfig[ConsulMiddlewareConfigConsulHostFieldName] + ":" + middlewareConfig[ConsulMiddlewareConfigConsulPortFieldName]
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consul.NewClient(apiClient)
	return &ConsulDiscoveryClient{
		Host:             middlewareConfig[ConsulMiddlewareConfigConsulHostFieldName],
		Port:             middlewareConfig[ConsulMiddlewareConfigConsulPortFieldName],
		client:           client,
		config:           consulConfig,
		middlewareConfig: middlewareConfig,
		registerConfig:   registerConfig,
		pushStopCh:       make(chan struct{}),
	}, err
}

func (c *ConsulDiscoveryClient) Register(instance *ServiceInstance) error {
	var (
		push                = c.registerConfig[ConsulRegisterConfigPushFieldName] != ""
		healthCheckPortStr  = c.registerConfig[ConsulRegisterConfigHealthcheckPortFieldName]
		serviceRegistration *api.AgentServiceRegistration
	)
	if !push && healthCheckPortStr == "" {
		//如果指定使用pull，但是没给健康检查端口，panic
		panic("")
	}

	consulMeta := make(map[string]string, 2)
	//编码protoc和tag到consul的meta data中，方便对端解出
	consulMeta[consulMetaDataServiceProtocFieldName] = string(instance.Protoc)
	consulMeta[consulMetaDataServiceExtraConfigFieldName] = instance.ExtraConfig

	if push {
		//如果主动汇报
		serviceRegistration = &api.AgentServiceRegistration{
			ID:      instance.InstanceId,
			Name:    instance.ServiceName,
			Address: instance.Host,
			Port:    instance.Port,
			Meta:    consulMeta,
			Check: &api.AgentServiceCheck{
				TTL: fmt.Sprintf("%vs", conf.DiscoveryMiddlewareCheckHeartBeatDuration.Seconds()),
			},
		}

		go func() {
			consulConfig := api.DefaultConfig()
			consulClient, err := api.NewClient(consulConfig)
			if err != nil {
				panic(err)
			}

			ticker := time.NewTicker(3 * time.Second)
			for {
				select {
				case <-ticker.C:
					if err = consulClient.Agent().UpdateTTL("service:"+instance.InstanceId, "healthy", "passing"); err != nil {
						klog.Errorf("Failed to update TTL: %v\n", err)
					} else {
						klog.Tracef("consul TTL updated successfully")
					}
				case <-c.pushStopCh:
					klog.Infof("consul TTL updated stopped")
					return
				}
			}
		}()
	} else {
		//如果让consul拉健康状态，则暴露端口
		serviceRegistration = &api.AgentServiceRegistration{
			ID:      instance.InstanceId,
			Name:    instance.ServiceName,
			Address: instance.Host,
			Port:    instance.Port,
			Meta:    consulMeta,
			Check: &api.AgentServiceCheck{
				//严重错误多长时间后删除
				DeregisterCriticalServiceAfter: "10s",
				HTTP:                           "http://" + instance.Host + ":" + healthCheckPortStr + "/health",
				Interval:                       fmt.Sprintf("%vs", conf.DiscoveryMiddlewareCheckHeartBeatDuration.Seconds()),
			},
		}

		c.httpServer = &http.Server{
			Addr: fmt.Sprintf(":" + healthCheckPortStr),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}),
		}

		go func() {
			if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				klog.Errorf("Error starting HTTP server for health check: %v", err)
			}
		}()
	}

	return c.client.Register(serviceRegistration)
}

func (c *ConsulDiscoveryClient) DeRegister(instanceId string) error {
	push := c.registerConfig[ConsulRegisterConfigPushFieldName] != ""

	if push {
		close(c.pushStopCh)
	} else if c.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()
		if err := c.httpServer.Shutdown(ctx); err != nil {
			klog.Errorf("Error stopping HTTP server for health check: %v", err)
		}
	}

	klog.Infof("[%v] try deRegister", instanceId)
	serviceRegistration := &api.AgentServiceRegistration{
		ID: instanceId,
	}
	return c.client.Deregister(serviceRegistration)
}

func (c *ConsulDiscoveryClient) DiscoverServices(serviceName string) []*ServiceInstance {
	instanceList, ok := c.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*ServiceInstance)
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	instanceList, ok = c.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*ServiceInstance)
	} else {
		go func() {
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					return
				}
				var instances []*ServiceInstance
				for _, entry := range v {
					instance := convertConsulAgentServiceToServiceInstance(entry.Service)
					if instance != nil {
						instances = append(instances, instance)
					}
				}
				c.instancesMap.Store(serviceName, instances)
			}
			defer plan.Stop()
			err := plan.Run(c.config.Address)
			if err != nil {
				klog.Errorf("DiscoverServices plan err:%v", err)
				return
			}
		}()
	}

	entries, _, err := c.client.Service(serviceName, "", false, nil)
	if err != nil {
		c.instancesMap.Store(serviceName, []*ServiceInstance{})
		klog.Error("Discover ServiceInstance Error!")
		return nil
	}
	instances := make([]*ServiceInstance, 0, len(entries))
	for _, entry := range entries {
		if entry.Checks.AggregatedStatus() == api.HealthPassing {
			instance := convertConsulAgentServiceToServiceInstance(entry.Service)
			if instance != nil {
				instances = append(instances, instance)
			}
		}
	}
	c.instancesMap.Store(serviceName, instances)
	return instances
}

func convertConsulAgentServiceToServiceInstance(agentService *api.AgentService) *ServiceInstance {
	if agentService.Meta == nil ||
		agentService.Meta[consulMetaDataServiceProtocFieldName] == "" {
		//agentService.Meta[consulMetaDataServiceExtraConfigFieldName] == "" {
		return nil
	}

	return &ServiceInstance{
		ServiceName: agentService.Service,
		InstanceId:  agentService.ID,
		ServiceServeConf: ServiceServeConf{
			Protoc: ProtocType(agentService.Meta[consulMetaDataServiceProtocFieldName]),
			Host:   agentService.Address,
			Port:   agentService.Port,
		},
		ExtraConfig: agentService.Meta[consulMetaDataServiceExtraConfigFieldName],
	}
}

var _ DiscoverClient = (*ConsulDiscoveryClient)(nil)
