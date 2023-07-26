package main

import (
	"supernova/executor/app"
	"supernova/pkg/conf"
	processor_plugin_shell "supernova/processor-plugin/processor-plugin-shell"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	HealthCheckPort = 20002
	GRPCServePort   = 20003
)

func main() {
	klog.SetLevel(klog.LevelTrace)
	//todo 这里根据配置，初始化scheduler
	cfg := conf.GetCommonConfig(conf.Dev)

	//var tag string
	//var healthPort string
	//var grpcPort int
	//
	//flag.StringVar(&tag, "tag", "", "")
	//flag.StringVar(&healthPort, "healthPort", "", "")
	//flag.IntVar(&grpcPort, "grpcPort", 0, "")

	shellExecutor := new(processor_plugin_shell.Shell)
	builder := app.NewExecutorBuilder()
	executor, err := builder.WithCustomTag("A").WithResourceTag("LargeMemory").
		WithInstanceID("instance-2").WithConsulDiscovery(cfg.ConsulConf, HealthCheckPort).
		WithProcessor(shellExecutor).WithGrpcServe("9.134.5.191", GRPCServePort).Build()

	if err != nil {
		panic(err)
	}

	if err := executor.Start(); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Hour)
}
