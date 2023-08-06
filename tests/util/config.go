package util

import (
	"strconv"
	"supernova/pkg/conf"
)

var (
	SimpleWebServerPort     = 9000
	SchedulerServePortStart = 8080
	SchedulerPort           = SchedulerServePortStart
	SchedulerAddress        = "http://9.134.5.191:" + strconv.Itoa(SchedulerPort)

	K8sSchedulerPort    = 5050
	K8sSchedulerAddress = "http://localhost:" + strconv.Itoa(K8sSchedulerPort)
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
		CollectorEndpointHost: "9.134.5.191",
		CollectorEndpointPort: "4317",
	}
)

var (
	GracefulStoppedExecutorLogPath = "/exiasun-cbs/codes/go/supernova/tests/log/graceful_stopped_executor.log"
	KilledSchedulerLogPath         = "/exiasun-cbs/codes/go/supernova/tests/log/killed_scheduler.log"
)
