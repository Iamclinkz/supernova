package discovery

import (
	"context"
	"fmt"
	"net/http"
	"supernova/pkg/conf"
	"sync"

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
}

func NewConsulMiddlewareConfig(consulHost, consulPort string) MiddlewareConfig {
	return MiddlewareConfig{
		ConsulMiddlewareConfigConsulHostFieldName: consulHost,
		ConsulMiddlewareConfigConsulPortFieldName: consulPort,
	}
}

func NewConsulRegisterConfig(healthCheckPort string) RegisterConfig {
	return RegisterConfig{
		ConsulRegisterConfigHealthcheckPortFieldName: healthCheckPort,
	}
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
	}, err
}

func (consulClient *ConsulDiscoveryClient) Register(instance *ServiceInstance) error {
	if consulClient.registerConfig[ConsulRegisterConfigHealthcheckPortFieldName] == "" {
		panic("")
	}

	healthCheckPort := consulClient.registerConfig[ConsulRegisterConfigHealthcheckPortFieldName]
	consulMeta := make(map[string]string, 2)
	//编码protoc和tag到consul的meta data中，方便对端解出
	consulMeta[consulMetaDataServiceProtocFieldName] = string(instance.Protoc)
	consulMeta[consulMetaDataServiceExtraConfigFieldName] = instance.ExtraConfig

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instance.InstanceId,
		Name:    instance.ServiceName,
		Address: instance.Host,
		Port:    instance.Port,
		Meta:    consulMeta,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + instance.Host + ":" + healthCheckPort + "/health",
			Interval:                       fmt.Sprintf("%vs", conf.DiscoveryMiddlewareCheckHeartBeatDuration.Seconds()),
		},
	}

	consulClient.httpServer = &http.Server{
		Addr: fmt.Sprintf(":" + healthCheckPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	go func() {
		if err := consulClient.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("Error starting HTTP server for health check: %v", err)
		}
	}()

	return consulClient.client.Register(serviceRegistration)
}

func (consulClient *ConsulDiscoveryClient) DeRegister(instanceId string) error {
	if consulClient.httpServer != nil {
		if err := consulClient.httpServer.Shutdown(context.Background()); err != nil {
			klog.Errorf("Error stopping HTTP server for health check: %v", err)
		}
	}

	serviceRegistration := &api.AgentServiceRegistration{
		ID: instanceId,
	}
	return consulClient.client.Deregister(serviceRegistration)
}

func (consulClient *ConsulDiscoveryClient) DiscoverServices(serviceName string) []*ServiceInstance {
	instanceList, ok := consulClient.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*ServiceInstance)
	}
	consulClient.mutex.Lock()
	defer consulClient.mutex.Unlock()
	instanceList, ok = consulClient.instancesMap.Load(serviceName)
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
				consulClient.instancesMap.Store(serviceName, instances)
			}
			defer plan.Stop()
			err := plan.Run(consulClient.config.Address)
			if err != nil {
				klog.Errorf("DiscoverServices plan err:%v", err)
				return
			}
		}()
	}

	entries, _, err := consulClient.client.Service(serviceName, "", false, nil)
	if err != nil {
		consulClient.instancesMap.Store(serviceName, []*ServiceInstance{})
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
	consulClient.instancesMap.Store(serviceName, instances)
	return instances
}

func convertConsulAgentServiceToServiceInstance(agentService *api.AgentService) *ServiceInstance {
	if agentService.Meta == nil ||
		agentService.Meta[consulMetaDataServiceProtocFieldName] == "" ||
		agentService.Meta[consulMetaDataServiceExtraConfigFieldName] == "" {
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
