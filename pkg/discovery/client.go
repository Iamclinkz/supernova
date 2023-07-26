package discovery

type Client interface {
	// Register 服务注册
	Register(instance *ServiceInstance) error
	//DeRegister 服务取消注册
	DeRegister(instanceId string) error
	//DiscoverServices 服务发现
	DiscoverServices(serviceName string) []*ServiceInstance
}
