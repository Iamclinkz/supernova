package otel_test

import (
	"log"
	"strconv"
	processor_plugin_idle "supernova/processor-plugin/processor-plugin-idle"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/tests/util"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func TestPureRun(t *testing.T) {
	var (
		LogLevel = klog.LevelWarn
	)

	supernovaTest := util.StartTest(3, 3, LogLevel, util.StartIdleExecutors,
		&processor_plugin_idle.IdleProcessorConfig{
			DoLog:    false,
			DoSleep:  true,
			SleepMin: 50 * time.Millisecond,
			SleepMax: 100 * time.Millisecond,
			DoFail:   false,
			FailRate: 0,
		})
	defer supernovaTest.EndTest()
	start := time.Now()

	//5000个Trigger，每个Trigger每隔5执行一次，相当于是每秒执行1000个trigger
	var triggerCount = 10000

	//加一个任务
	if err := util.RegisterJob(util.SchedulerAddress, &model.Job{
		Name:                  "test-idle-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Idle",
		GlueSource: map[string]string{
			"message": "hello from TestPureRun",
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
			ExecuteTimeout:    2 * time.Second,            //执行超过3s算超时。
			TriggerNextTime:   time.Now(),
			MisfireStrategy:   constance.MisfireStrategyTypeDoNothing,
			FailRetryInterval: 3 * time.Second, //重试间隔为1s
		}
	}

	util.RegisterTriggers(util.SchedulerAddress, triggers)

	log.Printf("register triggers successed, cost:%v\n", time.Since(start))
	//只是为了看看metrics上报，不结束
	time.Sleep(100 * time.Hour)
}
