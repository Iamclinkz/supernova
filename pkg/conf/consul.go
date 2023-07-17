package conf

var DevConsulConfig = &ConsulConf{
	Host: "localhost",
	Port: 8500,
}

var K8sConsulConfig = &ConsulConf{
	Host: "consul-service",
	Port: 8500,
}

type ConsulConf struct {
	Host string
	Port int
}
