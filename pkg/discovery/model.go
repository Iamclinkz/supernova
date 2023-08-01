package discovery

// ProtocType 服务提供的协议的类型（当前只支持http and grpc）
type ProtocType string

const (
	ProtocTypeGrpc ProtocType = "Grpc"
	ProtocTypeHttp ProtocType = "Http"
)

// ServiceServeConf 服务提供方指定，服务的协议、Host、端口
type ServiceServeConf struct {
	Protoc ProtocType
	Host   string
	Port   int
}

// ServiceInstance 需要被服务注册/发现的服务实例的原信息
type ServiceInstance struct {
	ServiceName      string //服务名称
	InstanceId       string //唯一标识Executor的InstanceID
	ServiceServeConf        //服务配置
	ExtraConfig      string //服务注册方和服务调用方约定好的额外信息
}

// MiddlewareType 使用的服务发现中间件
type MiddlewareType string

const (
	TypeConsul MiddlewareType = "Grpc"
	TypeK8s    MiddlewareType = "K8s"
)

type MiddlewareConfig map[string]string

type RegisterConfig map[string]string
