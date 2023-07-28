package discovery

// 用于约定服务发现中间件中的Executor实例中的元数据的Tag的前缀标识
const (
	TagPrefix         = "X-Tag-"
	EnvTagPrefix      = TagPrefix + "Env-"
	ResourceTagPrefix = TagPrefix + "Res-"
	GlueTypeTagPrefix = TagPrefix + "Glue-"
	CustomTagPrefix   = TagPrefix + "Custom-"
)
