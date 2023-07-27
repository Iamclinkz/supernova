package discovery

// ProtocType 服务提供的协议的类型（当前只支持http and grpc）
type ProtocType string

const (
	ProtocTypeGrpc ProtocType = "Grpc"
	ProtocTypeHttp ProtocType = "Http"
)

// ExecutorServiceServeConf 服务提供方指定，服务的协议、Host、端口
type ExecutorServiceServeConf struct {
	Protoc ProtocType
	Host   string
	Port   int
}

type ExecutorServiceInstance struct {
	InstanceId               string   //唯一标识Executor的InstanceID
	ExecutorServiceServeConf          //Executor服务配置
	Tags                     []string //Executor携带的Tags
}

// MiddlewareType 使用的服务发现中间件
type MiddlewareType string

const (
	TypeConsul MiddlewareType = "Grpc"
	TypeK8s                   = "K8s"
)

type MiddlewareConfig map[string]string

type RegisterConfig map[string]string
