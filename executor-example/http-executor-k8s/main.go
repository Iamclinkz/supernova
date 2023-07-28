package main

import (
	"flag"
	"supernova/executor/app"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// service config
var GrpcPort = flag.Int("grpcPort", 7070, "grpc port")

// log config
var LogLevel = flag.Int("logLevel", 4, "log level")

// discovery config
var K8sHealthCheckPort = flag.String("k8sHealthCheckPort", "9090", "health check port")

// namespace
var K8sNamespace = flag.String("k8sNamespace", "supernova", "k8s namespace")

func main() {
	flag.Parse()
	klog.SetLevel(klog.Level(*LogLevel))
	httpExecutor := new(processor_plugin_http.HTTP)
	builder := app.NewExecutorBuilder()
	executor, err := builder.
		WithK8sDiscovery(*K8sNamespace, *K8sHealthCheckPort).
		WithProcessor(httpExecutor).
		WithGrpcServe("0.0.0.0", *GrpcPort).
		Build()

	if err != nil {
		panic(err)
	}

	go executor.Start()

	time.Sleep(100 * 24 * time.Hour)
}
