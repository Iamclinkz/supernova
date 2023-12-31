package main

import (
	"flag"
	"strconv"
	"supernova/executor/app"
	"supernova/executor/processor"
	"supernova/pkg/session/trace"
	processor_plugin_shell "supernova/processor-plugin/processor-plugin-shell"

	"github.com/cloudwego/kitex/pkg/klog"
)

// service config
var GrpcHost = flag.String("grpcHost", "9.134.5.191", "grpc host")
var GrpcPort = flag.Int("grpcPort", 20001, "grpc port")

// log config
var LogLevel = flag.Int("logLevel", 4, "log level")

// discovery config
var HealthCheckPort = flag.Int("healthCheckPort", 10101, "health check port")
var ConsulHost = flag.String("consulHost", "9.134.5.191", "consul host")
var ConsulPort = flag.Int("consulPort", 8500, "consul port")

func main() {
	flag.Parse()
	klog.SetLevel(klog.LevelTrace)

	shellExecutor := new(processor_plugin_shell.Shell)
	builder := app.NewExecutorBuilder()
	executor, err := builder.
		WithCustomTag("A").
		WithResourceTag("LargeMemory").
		WithInstanceID("instance-2").
		WithConsulDiscovery(*ConsulHost, strconv.Itoa(*ConsulPort), *HealthCheckPort).
		WithProcessor(shellExecutor, &processor.ProcessConfig{
			Async:          false,
			MaxWorkerCount: 500,
		}).
		WithGrpcServe("9.134.5.191", *GrpcPort).
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
