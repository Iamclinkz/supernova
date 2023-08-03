package util

import (
	"fmt"
	"supernova/executor/app"
	"supernova/executor/processor"
	"supernova/pkg/session/trace"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"
	processor_plugin_idle "supernova/processor-plugin/processor-plugin-idle"

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

type ExecutorStartFunc func(instanceConfigs []*ExecutorInstanceConf, executorConfig any) []*app.Executor

func StartHttpExecutors(instanceConfigs []*ExecutorInstanceConf, extraConfig any) []*app.Executor {
	ret := make([]*app.Executor, 0, len(instanceConfigs))
	for _, instanceConf := range instanceConfigs {
		httpExecutor := new(processor_plugin_http.HTTP)
		builder := app.NewExecutorBuilder()
		executor, err := builder.WithInstanceID(GenUnderCloudExecutorID()).WithConsulDiscovery(
			DevConsulHost, DevConsulPort, instanceConf.HealthCheckPort).
			WithProcessor(httpExecutor, &processor.ProcessConfig{
				Async:          true,
				MaxWorkerCount: 10000,
			}).WithGrpcServe(instanceConf.GrpcServeHost, instanceConf.GrpcServePort).
			WithOTelConfig(&trace.OTelConfig{
				EnableTrace:    false,
				EnableMetrics:  true,
				InstrumentConf: DevTraceConfig,
			}).
			Build()

		if err != nil {
			panic(err)
		}

		ret = append(ret, executor)
		go executor.Start()
	}

	return ret
}

func StartIdleExecutors(instanceConfigs []*ExecutorInstanceConf,
	extraConfig any) []*app.Executor {
	idleConf := extraConfig.(*processor_plugin_idle.IdleProcessorConfig)
	ret := make([]*app.Executor, 0, len(instanceConfigs))
	for _, instanceConf := range instanceConfigs {
		idleExecutor := processor_plugin_idle.NewIdle(idleConf)
		builder := app.NewExecutorBuilder()
		executor, err := builder.WithInstanceID(GenUnderCloudExecutorID()).WithConsulDiscovery(
			DevConsulHost, DevConsulPort, instanceConf.HealthCheckPort).
			WithProcessor(idleExecutor, &processor.ProcessConfig{
				Async:          true,
				MaxWorkerCount: 10000,
			}).WithGrpcServe(instanceConf.GrpcServeHost, instanceConf.GrpcServePort).
			WithOTelConfig(&trace.OTelConfig{
				EnableTrace:    false,
				EnableMetrics:  true,
				InstrumentConf: DevTraceConfig,
			}).
			Build()

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
