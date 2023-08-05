package main

import (
	"flag"
	"os"
	"supernova/executor/app"
	"supernova/executor/processor"
	"supernova/pkg/session/trace"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"

	"github.com/cloudwego/kitex/pkg/klog"
)

type SetupConfig struct {
	GrpcPort           int
	LogLevel           int
	K8sHealthCheckPort string
	K8sNamespace       string
}

var setupConfig SetupConfig

func init() {
	flag.IntVar(&setupConfig.GrpcPort, "grpcPort", 7070, "grpc serve port")
	flag.IntVar(&setupConfig.LogLevel, "logLevel", 1, "log level")
	flag.StringVar(&setupConfig.K8sHealthCheckPort, "k8sHealthCheckPort", "9090", "health check port")
	flag.StringVar(&setupConfig.K8sNamespace, "k8sNamespace", "supernova", "k8s namespace")
	flag.Parse()
}

func main() {
	klog.Infof("use config:%v", setupConfig)
	klog.SetLevel(klog.Level(setupConfig.LogLevel))
	httpExecutor := new(processor_plugin_http.HTTP)
	builder := app.NewExecutorBuilder()
	executor, err := builder.
		WithK8sDiscovery(setupConfig.K8sNamespace, setupConfig.K8sHealthCheckPort).
		WithProcessor(httpExecutor, &processor.ProcessConfig{
			Async:          true,
			MaxWorkerCount: 10000,
		}).
		WithGrpcServe("0.0.0.0", setupConfig.GrpcPort).
		WithInstanceID(os.Getenv("HOSTNAME")).
		//todo
		WithOTelConfig(&trace.OTelConfig{
			EnableTrace:   false,
			EnableMetrics: false,
		}).
		Build()

	if err != nil {
		panic(err)
	}

	executor.Start()
}
