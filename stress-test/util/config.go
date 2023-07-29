package util

import (
	"strconv"
	"supernova/pkg/conf"
)

var (
	SimpleWebServerPort     = 9000
	SchedulerServePortStart = 8080
	SchedulerPort           = SchedulerServePortStart
	SchedulerAddress        = "http://localhost:" + strconv.Itoa(SchedulerPort)
)

var (
	//mysql
	DevMysqlConfig = &conf.MysqlConf{
		Host:               "localhost",
		Port:               "3306",
		UserName:           "root",
		Password:           "password",
		DbName:             "supernova",
		MaxIdleConnections: 16,
		MaxOpenConnections: 128,
	}

	//consul
	DevConsulHost = "9.134.5.191"
	DevConsulPort = "8500"

	//trace
	DevTraceConfig = &conf.OTelConf{
		ExportEndpointHost: "9.134.5.191",
		ExportEndpointPort: "4317",
	}
)
