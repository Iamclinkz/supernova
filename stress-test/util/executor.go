package util

import (
	"fmt"
	"supernova/executor/app"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"

	"github.com/google/uuid"
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

func StartHttpExecutors(instanceConfigs []*ExecutorInstanceConf) []*app.Executor {
	ret := make([]*app.Executor, 0, len(instanceConfigs))
	for _, instanceConf := range instanceConfigs {
		httpExecutor := new(processor_plugin_http.HTTP)
		builder := app.NewExecutorBuilder()
		executor, err := builder.WithInstanceID(GenUnderCloudExecutorID()).WithConsulDiscovery(
			DevConsulHost, DevConsulPort, instanceConf.HealthCheckPort).
			WithProcessor(httpExecutor).WithGrpcServe(instanceConf.GrpcServeHost, instanceConf.GrpcServePort).Build()

		if err != nil {
			panic(err)
		}

		ret = append(ret, executor)
		go executor.Start()
	}

	return ret
}

// GenUnderCloudExecutorID 为云下测试生成一个唯一的ExecutorID（云上直接用pod名，不会重复）
func GenUnderCloudExecutorID() string {
	return fmt.Sprintf("Executor-%v", uuid.New())
}
