package util

import (
	"strconv"
	"supernova/pkg/conf"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"

	"github.com/cloudwego/kitex/pkg/klog"
)

var (
	SchedulerServePortStart = 8080
)

func StartScheduler(port string) {
	cfg := conf.GetCommonConfig(conf.Dev)

	builder := app.NewSchedulerBuilder()
	scheduler, err := builder.WithMysqlStore(cfg.MysqlConf).WithConsulDiscovery(cfg.ConsulConf).Build()
	scheduler.Start()

	if err != nil {
		panic(err)
	}

	router := http.InitHttpHandler(scheduler)
	klog.Infof("Start the server at %v", port)

	go func() {
		if err = router.Run(":" + port); err != nil {
			panic(err)
		}
	}()
}

func StartSchedulers(count int) {
	cfg := conf.GetCommonConfig(conf.Dev)

	for i := 1; i <= count; i++ {
		builder := app.NewSchedulerBuilder()
		scheduler, err := builder.WithMysqlStore(cfg.MysqlConf).WithConsulDiscovery(cfg.ConsulConf).Build()
		scheduler.Start()

		if err != nil {
			panic(err)
		}

		router := http.InitHttpHandler(scheduler)
		klog.Infof("Start the server at %v", SchedulerServePortStart+count)

		go func(port int) {
			if err = router.Run(":" + strconv.Itoa(port)); err != nil {
				panic(err)
			}
		}(SchedulerServePortStart + i)
	}
}
