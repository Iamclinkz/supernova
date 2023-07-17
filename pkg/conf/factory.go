package conf

type Env string

const (
	Dev Env = "dev"
	K8s Env = "k8s"
)

type CommonConf struct {
	*OTelConf
	*RedisConf
	*MysqlConf
	*KongAdminConf
	*ConsulConf
}

func GetCommonConfig(env Env) *CommonConf {
	switch env {
	case K8s:
		return &CommonConf{
			OTelConf:      K8sTraceConfig,
			RedisConf:     K8sRedisConfig,
			MysqlConf:     K8sMysqlConfig,
			KongAdminConf: K8sKongAdminConfig,
			ConsulConf:    K8sConsulConfig,
		}
	default:
		return &CommonConf{
			OTelConf:      DevTraceConfig,
			RedisConf:     DevRedisConfig,
			MysqlConf:     DevMysqlConfig,
			KongAdminConf: DevKongAdminConfig,
			ConsulConf:    DevConsulConfig,
		}
	}
}
