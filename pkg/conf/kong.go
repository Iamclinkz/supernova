package conf

var DevKongAdminConfig = &KongAdminConf{
	Host:         "localhost",
	Port:         "8001",
	ConsumerName: "user",
}

var K8sKongAdminConfig = &KongAdminConf{
	Host:         "kong-service",
	Port:         "8001",
	ConsumerName: "user",
}

type KongAdminConf struct {
	Host         string
	Port         string
	ConsumerName string
}
