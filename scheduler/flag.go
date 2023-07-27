package main

import (
	"flag"
	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"supernova/pkg/conf"
)

type SetupConfig struct {
	HttpPort      string
	LogLevel      int
	DiscoveryType string
	ConsulHost    string
	ConsulPort    string
	DBType        string
	MysqlConf     conf.MysqlConf
	OTelConf      conf.OTelConf
}

var setupConfig SetupConfig

func init() {
	flag.StringVar(&setupConfig.HttpPort, "httpPort", "8080", "http port")
	flag.IntVar(&setupConfig.LogLevel, "logLevel", 1, "log level")
	flag.StringVar(&setupConfig.DiscoveryType, "discoveryType", "consul", "discovery type")
	flag.StringVar(&setupConfig.ConsulHost, "consulHost", "9.134.5.191", "consul host")
	flag.StringVar(&setupConfig.ConsulPort, "consulPort", "8500", "consul port")
	flag.StringVar(&setupConfig.DBType, "dbType", "mysql", "db type, mysql or mongo")
	flag.StringVar(&setupConfig.MysqlConf.Host, "mysqlHost", "9.134.5.191", "MySQL host")
	flag.StringVar(&setupConfig.MysqlConf.Port, "mysqlPort", "3306", "MySQL port")
	flag.StringVar(&setupConfig.MysqlConf.UserName, "mysqlUsername", "root", "MySQL username")
	flag.StringVar(&setupConfig.MysqlConf.Password, "mysqlPassword", "password", "MySQL password")
	flag.StringVar(&setupConfig.MysqlConf.DbName, "mysqlDbname", "supernova", "MySQL database name")
	flag.IntVar(&setupConfig.MysqlConf.MaxIdleConnections, "mysqlMaxIdleConn", 16, "MySQL max idle connections")
	flag.IntVar(&setupConfig.MysqlConf.MaxOpenConnections, "mysqlMaxOpenConn", 128, "MySQL max open connections")
	flag.Parse()
	log.Infof("find configs: %+v", setupConfig)
}