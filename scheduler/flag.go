package main

import (
	"flag"
	"os"
	"supernova/pkg/conf"
	"supernova/pkg/util"

	"github.com/cloudwego/kitex/tool/internal_pkg/log"
)

type SetupConfig struct {
	InstanceID    string
	HttpPort      string
	LogLevel      int
	DiscoveryType string
	ConsulHost    string
	ConsulPort    string
	DBType        string
	MysqlConf     conf.MysqlConf
	OTelConf      conf.OTelConf
	K8sNamespace  string
}

var setupConfig SetupConfig

func init() {
	flag.StringVar(&setupConfig.InstanceID, "instanceID", "instance-unknown", "instance id")
	flag.StringVar(&setupConfig.HttpPort, "httpPort", "8080", "http port")
	flag.IntVar(&setupConfig.LogLevel, "logLevel", 2, "log level")
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
	flag.StringVar(&setupConfig.K8sNamespace, "k8sNamespace", "supernova", "k8s namespace")
	flag.StringVar(&setupConfig.OTelConf.CollectorEndpointHost, "otelEndpointHost", "9.134.5.191", "otelCollectorHost")
	flag.StringVar(&setupConfig.OTelConf.CollectorEndpointPort, "otelEndpointPort", "4317", "otelCollectorPort")
	flag.Parse()
	if util.GetEnv() == "k8s" {
		//来不及了，先这样吧
		setupConfig.InstanceID = os.Getenv("HOSTNAME")
	}
	log.Infof("find configs: %+v", setupConfig)
}
