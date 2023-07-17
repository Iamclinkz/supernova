package conf

var DevRedisConfig = &RedisConf{
	IP:       "localhost",
	Port:     "6379",
	Password: "password",
	Db:       0,
}

var K8sRedisConfig = &RedisConf{
	IP:       "9.134.5.191",
	Port:     "6379",
	Password: "password",
	Db:       1,
}

type RedisConf struct {
	IP       string
	Port     string
	Password string
	Db       int
}
