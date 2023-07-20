package discovery

import "errors"

type MiddlewareType string

const (
	TypeConsul MiddlewareType = "Grpc"
	TypeMysql                 = "Http"
	TypeK8s                   = "K8s"
)

type MiddlewareConfig struct {
	Type MiddlewareType
	Host string
	Port int
}

func NewDiscoveryClient(config *MiddlewareConfig) (Client, error) {
	switch config.Type {
	case TypeConsul:
		return newConsulDiscoverClient(config.Host, config.Port)
	default:
		return nil, errors.New("no such discovery client")
	}
}
