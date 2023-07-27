package conf

type MysqlConf struct {
	Host               string
	Port               string
	UserName           string
	Password           string
	DbName             string
	MaxIdleConnections int
	MaxOpenConnections int
}
