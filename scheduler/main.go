package main

import (
	"flag"
	"strconv"
	"supernova/pkg/conf"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"

	"github.com/cloudwego/kitex/pkg/klog"
)

var port = flag.Int("port", 8080, "http server port")
var logLevel = flag.Int("logLevel", 4, "log level")

func main() {
	flag.Parse()
	klog.SetLevel(klog.Level(*logLevel))
	
	//todo 这里根据配置，初始化scheduler
	cfg := conf.GetCommonConfig(conf.Dev)

	//指定依赖，组装Scheduler
	builder := app.NewSchedulerBuilder()
	scheduler, err := builder.WithMysqlStore(cfg.MysqlConf).WithConsulDiscovery(cfg.ConsulConf).Build()
	if err != nil {
		panic(err)
	}
	scheduler.Start()

	if err != nil {
		panic(err)
	}

	router := http.InitHttpHandler(scheduler)
	klog.Infof("Start the server at %v", *port)
	if err = router.Run(":" + strconv.Itoa(*port)); err != nil {
		klog.Fatalf("failed to start HTTP server: %v", err)
	}
}
