package discovery

import "errors"

type Type string

const (
	TypeConsul Type = "Grpc"
	TypeMysql       = "Http"
	TypeK8s         = "K8s"
)

type Config struct {
	Host string
	Port int
}

func NewDiscoveryClient(t Type, config *Config) (Client, error) {
	switch t {
	case TypeConsul:
		return newConsulDiscoverClient(config.Host, config.Port)
	default:
		return nil, errors.New("no such discovery client")
	}
}
