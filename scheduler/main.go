package main

import (
	"supernova/pkg/conf"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"

	"github.com/cloudwego/kitex/pkg/klog"
)

func main() {
	klog.SetLevel(klog.LevelTrace)
	//todo 这里根据配置，初始化scheduler
	cfg := conf.GetCommonConfig(conf.Dev)

	builder := app.NewSchedulerBuilder()
	scheduler, err := builder.WithMysqlStore(cfg.MysqlConf).WithConsulDiscovery(cfg.ConsulConf).Build()
	scheduler.Start()

	if err != nil {
		panic(err)
	}

	router := http.InitHttpHandler(scheduler)
	klog.Infof("Start the server at %v", 8080)
	if err = router.Run(":8080"); err != nil {
		klog.Fatalf("failed to start HTTP server: %v", err)
	}
}
