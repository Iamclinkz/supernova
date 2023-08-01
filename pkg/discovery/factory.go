package discovery

type DiscoverClient interface {
	// Register 服务注册
	Register(instance *ServiceInstance) error
	//DeRegister 服务取消注册
	DeRegister(instanceId string) error
	//DiscoverServices 服务发现
	DiscoverServices(serviceName string) []*ServiceInstance
}

func NewDiscoveryClient(t MiddlewareType, middlewareConfig MiddlewareConfig,
	registerConfig RegisterConfig) (DiscoverClient, error) {
	switch t {
	case TypeConsul:
		return newConsulDiscoveryClient(middlewareConfig, registerConfig)
	case TypeK8s:
		return newK8sDiscoveryClient(middlewareConfig, registerConfig)
	default:
		panic("")
	}
}
