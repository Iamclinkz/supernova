package conf

var DevMysqlConfig = &MysqlConf{
	Host:               "localhost",
	Port:               "3306",
	UserName:           "root",
	Password:           "password",
	DbName:             "supernova",
	MaxIdleConnections: 10,
	MaxOpenConnections: 100,
}

var K8sMysqlConfig = &MysqlConf{
	Host:               "9.134.5.191",
	Port:               "3306",
	UserName:           "root",
	Password:           "password",
	DbName:             "supernovaK8s",
	MaxIdleConnections: 10,
	MaxOpenConnections: 128,
}

type MysqlConf struct {
	Host               string
	Port               string
	UserName           string
	Password           string
	DbName             string
	MaxIdleConnections int
	MaxOpenConnections int
}
