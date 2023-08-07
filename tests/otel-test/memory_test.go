package otel_test

import (
	"log"
	"math/rand"
	"strconv"
	"supernova/pkg/session/trace"
	processor_plugin_idle "supernova/processor-plugin/processor-plugin-idle"
	"supernova/scheduler/app"
	"supernova/scheduler/constance"
	"supernova/scheduler/handler/http"
	"supernova/scheduler/model"
	"supernova/tests/util"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func TestMemoryPureRun(t *testing.T) {
	var (
		LogLevel                 = klog.LevelWarn
		MemoryStoreSchedulerPort = "8180"
	)

	supernovaTest := util.StartTest(0, 4, LogLevel, util.StartIdleExecutors,
		&processor_plugin_idle.IdleProcessorConfig{
			DoLog:    false,
			DoSleep:  false,
			SleepMin: 50 * time.Millisecond,
			SleepMax: 100 * time.Millisecond,
			DoFail:   false,
			FailRate: 0.0,
		})
	defer supernovaTest.EndTest()
	start := time.Now()

	//5000个Trigger，每个Trigger每隔5执行一次，相当于是每秒执行1000个trigger
	var triggerCount = 200000

	builder := app.NewSchedulerBuilder()
	memoryStoreScheduler, err := builder.WithMemoryStore().
		WithConsulDiscovery(util.DevConsulHost, util.DevConsulPort).WithOTelConfig(&trace.OTelConfig{
		EnableTrace:    false,
		EnableMetrics:  true,
		InstrumentConf: util.DevTraceConfig,
	}).WithStandalone().Build()

	if err != nil {
		panic(err)
	}
	go memoryStoreScheduler.Start()
	router := http.InitHttpHandler(memoryStoreScheduler)
	klog.Infof("Start the server at %v", MemoryStoreSchedulerPort)

	go func() {
		if err = router.Run(":" + MemoryStoreSchedulerPort); err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)

	//加一个任务
	if err := util.RegisterJob("http://localhost:"+MemoryStoreSchedulerPort, &model.Job{
		Name:                  "test-idle-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Idle",
		GlueSource: map[string]string{
			"message": "hello",
		},
	}); err != nil {
		panic(err)
	}

	triggers := make([]*model.Trigger, triggerCount)

	for i := 0; i < triggerCount; i++ {
		triggers[i] = &model.Trigger{
			Name:              "test-trigger-" + strconv.Itoa(i),
			JobID:             1,
			ScheduleType:      constance.ScheduleTypeCron, //使用cron循环执行
			ScheduleConf:      "*/5 * * * * *",            //每5s执行一次
			FailRetryCount:    5,                          //最大失败重试五次。
			ExecuteTimeout:    5 * time.Second,            //执行超过5s算超时。
			TriggerNextTime:   time.Now().Add(time.Duration(rand.Intn(10000)) * time.Millisecond),
			MisfireStrategy:   constance.MisfireStrategyTypeDoNothing,
			FailRetryInterval: 3 * time.Second, //重试间隔为3s
		}
	}

	util.RegisterTriggers("http://localhost:"+MemoryStoreSchedulerPort, triggers)

	log.Printf("register triggers successed, cost:%v\n", time.Since(start))
	//只是为了看看metrics上报，不结束
	time.Sleep(100 * time.Hour)
}
