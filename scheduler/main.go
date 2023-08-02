package main

import (
	"strings"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"

	"github.com/cloudwego/kitex/pkg/klog"
)

func main() {
	klog.SetLevel(klog.Level(setupConfig.LogLevel))

	klog.Infof("init with setupConfig:%+v", setupConfig)
	builder := app.NewSchedulerBuilder()

	//db
	switch strings.ToLower(setupConfig.DBType) {
	case "mysql":
		builder.WithMysqlStore(&setupConfig.MysqlConf)
	}

	//discovery
	switch strings.ToLower(setupConfig.DiscoveryType) {
	case "consul":
		builder.WithConsulDiscovery(setupConfig.ConsulHost, setupConfig.ConsulPort)
	case "k8s":
		builder.WithK8sDiscovery(setupConfig.K8sNamespace)
	}

	scheduler, err := builder.Build()
	if err != nil {
		panic(err)
	}
	go scheduler.Start()

	router := http.InitHttpHandler(scheduler)
	klog.Infof("Start the server at %v", setupConfig.HttpPort)
	if err = router.Run(":" + setupConfig.HttpPort); err != nil {
		klog.Fatalf("failed to start HTTP server: %v", err)
	}
}
