package discovery

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type ConsulDiscoverClient struct {
	Host         string // TypeConsul Host
	Port         int    // TypeConsul Port
	client       consul.Client
	config       *api.Config
	mutex        sync.Mutex
	instancesMap sync.Map
	httpServer   *http.Server
}

func newConsulDiscoverClient(consulHost string, consulPort int) (Client, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(consulPort)
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consul.NewClient(apiClient)
	return &ConsulDiscoverClient{
		Host:   consulHost,
		Port:   consulPort,
		config: consulConfig,
		client: client,
	}, err
}

func (consulClient *ConsulDiscoverClient) Register(instance *ServiceInstance) error {
	if instance.Meta == nil {
		instance.Meta = make(map[string]string)
	}

	//编码protoc到meta中
	instance.Meta[serviceProtocFieldName] = string(instance.Protoc)

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instance.InstanceId,
		Name:    instance.ServiceName,
		Address: instance.Host,
		Port:    instance.Port,
		Meta:    instance.Meta,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           instance.MiddlewareHealthCheckUrl,
			Interval:                       "15s",
		},
	}

	consulClient.httpServer = &http.Server{
		Addr: fmt.Sprintf(":" + strconv.Itoa(instance.Port)),
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

func (consulClient *ConsulDiscoverClient) DeRegister(instanceId string) error {
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

func (consulClient *ConsulDiscoverClient) DiscoverServices(serviceName string) []*ServiceInstance {
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
					instance := convertAgentServiceToServiceInstance(entry.Service)
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
			instance := convertAgentServiceToServiceInstance(entry.Service)
			if instance != nil {
				instances = append(instances, instance)
			}
		}
	}
	consulClient.instancesMap.Store(serviceName, instances)
	return instances
}

func convertAgentServiceToServiceInstance(agentService *api.AgentService) *ServiceInstance {
	if agentService.Meta == nil || agentService.Meta[serviceProtocFieldName] == "" {
		return nil
	}

	return &ServiceInstance{
		ServiceName: agentService.Service,
		InstanceId:  agentService.ID,
		ServiceServeConf: ServiceServeConf{
			Protoc: ProtocType(agentService.Meta[serviceProtocFieldName]),
			Host:   agentService.Address,
			Port:   agentService.Port,
		},
		Meta: agentService.Meta,
	}
}

var _ Client = (*ConsulDiscoverClient)(nil)
