package main

import (
	"flag"
	"strconv"
	"supernova/executor/app"
	"supernova/executor/processor"
	"supernova/pkg/session/trace"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// service config
var GrpcHost = flag.String("grpcHost", "9.134.5.191", "grpc host")
var GrpcPort = flag.Int("grpcPort", 20001, "grpc port")

// log config
var LogLevel = flag.Int("logLevel", 5, "log level")

// discovery config
var HealthCheckPort = flag.Int("healthCheckPort", 9090, "health check port")
var ConsulHost = flag.String("consulHost", "9.134.5.191", "consul host")
var ConsulPort = flag.Int("consulPort", 8500, "consul port")

func main() {
	flag.Parse()
	klog.SetLevel(klog.Level(*LogLevel))
	httpExecutor := new(processor_plugin_http.HTTP)
	builder := app.NewExecutorBuilder()
	executor, err := builder.
		WithCustomTag("A").
		WithResourceTag("LargeMemory").
		WithInstanceID("instance-1").
		WithConsulDiscovery(*ConsulHost, strconv.Itoa(*ConsulPort), *HealthCheckPort).
		WithProcessor(httpExecutor, &processor.ProcessConfig{
			Async:          true,
			MaxWorkerCount: 10000,
		}).
		WithGrpcServe(*GrpcHost, *GrpcPort).
		WithOTelConfig(&trace.OTelConfig{
			EnableTrace:   false,
			EnableMetrics: false,
		}).
		Build()

	if err != nil {
		panic(err)
	}

	go executor.Start()

	time.Sleep(12 * time.Hour)
}
