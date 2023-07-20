package discovery

const (
	serviceProtocFieldName = "X-Protoc-Type"
)

type ProtocType string

const (
	ProtocTypeGrpc ProtocType = "Grpc"
	ProtocTypeHttp            = "Http"
)

type ServiceServeConf struct {
	Protoc ProtocType
	Host   string
	Port   int
}

type ServiceInstance struct {
	ServiceName string
	InstanceId  string
	//如果使用的中间件检查健康，例如consul，那么应该填写这个字段，让consul检查
	MiddlewareHealthCheckUrl string
	ServiceServeConf
	Meta map[string]string
}
