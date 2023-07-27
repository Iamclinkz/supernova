package util

import (
	"fmt"
	"supernova/executor/app"
	"supernova/pkg/conf"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"
)

type ExecutorInstanceConf struct {
	InstanceID      string
	GrpcServeHost   string
	GrpcServePort   int
	HealthCheckPort int
}

const (
	InstancePrefix       = "Test-Instance-"
	GrpcServeHost        = "9.134.5.191"
	GrpcServePortStart   = 10000
	HealthCheckPortStart = 11000
)

func GenExecutorInstanceConfWithCount(count int) []*ExecutorInstanceConf {
	ret := make([]*ExecutorInstanceConf, count)

	for i := 1; i <= count; i++ {
		ret[i-1] = GenExecutorInstanceConf(i)
	}
	return ret
}

func GenExecutorInstanceConf(id int) *ExecutorInstanceConf {
	return &ExecutorInstanceConf{
		InstanceID:      fmt.Sprintf("%s%d", InstancePrefix, id),
		GrpcServeHost:   GrpcServeHost,
		GrpcServePort:   GrpcServePortStart + id,
		HealthCheckPort: HealthCheckPortStart + id,
	}
}

func StartHttpExecutors(instanceConfigs []*ExecutorInstanceConf) {
	for _, instanceConf := range instanceConfigs {
		cfg := conf.GetCommonConfig(conf.Dev)
		httpExecutor := new(processor_plugin_http.HTTP)
		builder := app.NewExecutorBuilder()
		executor, err := builder.WithInstanceID(instanceConf.InstanceID).WithConsulDiscovery(cfg.ConsulConf, instanceConf.HealthCheckPort).
			WithProcessor(httpExecutor).WithGrpcServe(instanceConf.GrpcServeHost, instanceConf.GrpcServePort).Build()

		if err != nil {
			panic(err)
		}

		go executor.Start()
	}
}
