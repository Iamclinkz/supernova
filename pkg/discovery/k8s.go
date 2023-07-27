package discovery

type K8sDiscoveryClient struct {
}

func newK8sDiscoveryClient(middlewareConfig map[string]string) (*K8sDiscoveryClient, error) {
	return nil, nil
}

func (k K8sDiscoveryClient) Register(instance *ExecutorServiceInstance, extraConfig map[string]string) error {
	//TODO implement me
	panic("implement me")
}

func (k K8sDiscoveryClient) DeRegister(instanceId string) error {
	//TODO implement me
	panic("implement me")
}

func (k K8sDiscoveryClient) DiscoverServices() []*ExecutorServiceInstance {
	//TODO implement me
	panic("implement me")
}

var _ ExecutorDiscoveryClient = (*K8sDiscoveryClient)(nil)
