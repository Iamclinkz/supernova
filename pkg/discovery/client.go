package discovery

type Client interface {
	Register(instance *ServiceInstance) error
	DeRegister(instanceId string) error
	DiscoverServices(serviceName string) []*ServiceInstance
}
