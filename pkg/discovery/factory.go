package discovery

type ExecutorDiscoveryClient interface {
	// Register 服务注册
	Register(instance *ExecutorServiceInstance, extraConfig map[string]string) error
	//DeRegister 服务取消注册
	DeRegister(instanceId string) error
	//DiscoverServices 服务发现
	DiscoverServices() []*ExecutorServiceInstance
}

func NewDiscoveryClient(t MiddlewareType, middlewareConfig map[string]string) (ExecutorDiscoveryClient, error) {
	switch t {
	case TypeConsul:
		return newConsulDiscoveryClient(middlewareConfig)
	case TypeK8s:
		return newK8sDiscoveryClient(middlewareConfig)
	default:
		panic("")
	}
}
