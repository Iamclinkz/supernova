package discovery

type Client interface {
	//服务注册
	Register(instance *ServiceInstance) error
	//服务取消注册
	DeRegister(instanceId string) error
	//服务发现
	DiscoverServices(serviceName string) []*ServiceInstance
}
