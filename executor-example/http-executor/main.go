package main

import (
	"flag"
	"supernova/executor/app"
	"supernova/pkg/conf"
	processor_plugin_http "supernova/processor-plugin/processor-plugin-http"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// service config
var GrpcHost = flag.String("grpcHost", "9.134.5.191", "grpc host")
var GrpcPort = flag.Int("grpcPort", 20001, "grpc port")

// log config
var LogLevel = flag.Int("logLevel", 4, "log level")

// discovery config
var HealthCheckPort = flag.Int("healthCheckPort", 8080, "health check port")
var ConsulHost = flag.String("consulHost", "9.134.5.191", "consul host")
var ConsulPort = flag.Int("consulPort", 8500, "consul port")

func main() {
	klog.SetLevel(klog.Level(*LogLevel))
	httpExecutor := new(processor_plugin_http.HTTP)
	builder := app.NewExecutorBuilder()
	executor, err := builder.WithCustomTag("A").WithResourceTag("LargeMemory").
		WithInstanceID("instance-1").WithConsulDiscovery(&conf.ConsulConf{
		Host: *ConsulHost,
		Port: *ConsulPort,
	}, *HealthCheckPort).WithProcessor(httpExecutor).WithGrpcServe(*GrpcHost, *GrpcPort).Build()

	if err != nil {
		panic(err)
	}

	go executor.Start()

	time.Sleep(12 * time.Hour)
}
