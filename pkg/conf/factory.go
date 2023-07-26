package conf

type Env string

const (
	Dev Env = "dev"
	K8s Env = "k8s"
)

type CommonConf struct {
	*OTelConf
	*MysqlConf
	*ConsulConf
}

func GetCommonConfig(env Env) *CommonConf {
	switch env {
	case K8s:
		return &CommonConf{
			OTelConf:   K8sTraceConfig,
			MysqlConf:  K8sMysqlConfig,
			ConsulConf: K8sConsulConfig,
		}
	default:
		return &CommonConf{
			OTelConf:   DevTraceConfig,
			MysqlConf:  DevMysqlConfig,
			ConsulConf: DevConsulConfig,
		}
	}
}
