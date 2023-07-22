package main

import (
	"log"
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/stress-test/util"
	"testing"
	"time"
)

var (
	SimpleWebServerPort = 9000
	SchedulerPort       = util.SchedulerServePortStart + 1
	SchedulerAddress    = "http://localhost:" + strconv.Itoa(SchedulerPort)
)

func init() {
	util.StartHttpExecutors(util.GenExecutorInstanceConfWithCount(3))
	util.StartSchedulers(2)
}

func TestBasic(t *testing.T) {
	httpServer := util.NewSimpleHttpServer(&util.SimpleHttpServerConf{
		FailRate:      0,
		ListeningPort: SimpleWebServerPort,
	})
	httpServer.Start()

	//加一个任务
	if err := util.RegisterJob(SchedulerAddress, &model.Job{
		Name:                  "test-http-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Http",
		GlueSource: map[string]string{
			"method":     "GET",
			"url":        "localhost:" + strconv.Itoa(SimpleWebServerPort),
			"timeout":    "30",
			"expectCode": "200",
			"expectBody": "ok",
			"debug":      "",
		},
	}); err != nil {
		panic(err)
	}

	for i := 0; i < 3; i++ {
		if err := util.RegisterTrigger(SchedulerAddress, &model.Trigger{
			Name:            "test-trigger-" + strconv.Itoa(i),
			JobID:           1,
			ScheduleType:    2,          //执行一次
			FailRetryCount:  0,          //失败不执行
			ExecuteTimeout:  2000000000, //2s
			TriggerNextTime: time.Now(),
			Param: map[string]string{
				util.ExecutorIDFieldName: strconv.Itoa(i),
			},
		}); err != nil {
			panic(err)
		}
	}

	for {
		time.Sleep(2 * time.Second)
		log.Println(httpServer.GetResult())
	}
}
