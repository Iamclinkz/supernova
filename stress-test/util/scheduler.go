package util

import (
	"strconv"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func StartSchedulers(count int) {
	for i := 1; i <= count; i++ {
		builder := app.NewSchedulerBuilder()
		scheduler, err := builder.WithMysqlStore(DevMysqlConfig).WithConsulDiscovery(DevConsulHost, DevConsulPort).Build()
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
		time.Sleep(1 * time.Second)
	}
}
