package discovery

const (
	serviceProtocFieldName = "X-Protoc-Type"
)

type ProtocType string

const (
	ProtocTypeGrpc ProtocType = "Grpc"
	ProtocTypeHttp ProtocType = "Http"
)

type ServiceServeConf struct {
	Protoc ProtocType
	Host   string
	Port   int
}

type ServiceInstance struct {
	ServiceName string
	InstanceId  string
	ServiceServeConf
	Meta map[string]string
}
