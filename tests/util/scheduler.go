package util

import (
	"strconv"
	"supernova/scheduler/app"
	"supernova/scheduler/handler/http"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func StartSchedulers(count int) []*app.Scheduler {
	ret := make([]*app.Scheduler, 0, count)
	for i := 0; i < count; i++ {
		builder := app.NewSchedulerBuilder()
		scheduler, err := builder.WithMysqlStore(DevMysqlConfig).
			WithConsulDiscovery(DevConsulHost, DevConsulPort).
			WithOTelCollector(DevTraceConfig).
			WithInstanceID("Test-Scheduler-" + strconv.Itoa(i)).
			Build()

		if err != nil {
			panic(err)
		}

		ret = append(ret, scheduler)
		go scheduler.Start()

		router := http.InitHttpHandler(scheduler)
		klog.Infof("Start the server at %v", SchedulerServePortStart+i)

		go func(port int) {
			if err = router.Run(":" + strconv.Itoa(port)); err != nil {
				panic(err)
			}
		}(SchedulerServePortStart + i)
		time.Sleep(1 * time.Second)
	}

	return ret
}
