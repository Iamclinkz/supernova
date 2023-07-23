package main

import (
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/stress-test/util"
	"testing"
	"time"

	"log"
)

var (
	SimpleWebServerPort = 9000
	SchedulerPort       = util.SchedulerServePortStart + 1
	SchedulerAddress    = "http://localhost:" + strconv.Itoa(SchedulerPort)
)

func initTest() {
	util.StartHttpExecutors(util.GenExecutorInstanceConfWithCount(3))
	util.StartSchedulers(2)
}

// TestWithoutFail 执行单次，任务不会失败
func TestWithoutFail(t *testing.T) {
	start := time.Now()

	var triggerCount = 5000

	initTest()
	httpServer := util.NewSimpleHttpServer(&util.SimpleHttpServerConf{
		FailRate:             0,
		ListeningPort:        SimpleWebServerPort,
		TriggerCount:         triggerCount,
		AllowDuplicateCalled: false,
	})
	go httpServer.Start()

	time.Sleep(5 * time.Second)
	//加一个任务
	if err := util.RegisterJob(SchedulerAddress, &model.Job{
		Name:                  "test-http-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Http",
		GlueSource: map[string]string{
			"method":     "POST",
			"url":        "http://localhost:" + strconv.Itoa(SimpleWebServerPort) + "/test",
			"timeout":    "30",
			"expectCode": "200",
			"expectBody": "OK",
			"debug":      "",
		},
	}); err != nil {
		panic(err)
	}

	triggers := make([]*model.Trigger, triggerCount)

	for i := 0; i < triggerCount; i++ {
		triggers[i] = &model.Trigger{
			Name:            "test-trigger-" + strconv.Itoa(i),
			JobID:           1,
			ScheduleType:    2,          //执行一次
			FailRetryCount:  5,          //失败重试五次。因为simple_http_service每次都需要开go程执行请求，瞬间很多个请求打过去可能造成失败的情况。。
			ExecuteTimeout:  3000000000, //3s
			TriggerNextTime: time.Now(),
			MisfireStrategy: constance.MisfireStrategyTypeDoNothing,
			Param: map[string]string{
				//把triggerID带过去，由SimpleHttpServer记录执行情况
				util.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}

	util.RegisterTriggers(SchedulerAddress, triggers)

	log.Printf("register triggers successed, cost:%v\n", time.Since(start))

	time.Sleep(15 * time.Second)

	httpServer.CheckResult(&util.CheckResultConf{
		AllSuccess:                    true,
		NoUncalledTriggers:            true,
		FailTriggerRateNotGreaterThan: 0.1,
	})
}

// TestRandomFailJob 执行单次，任务有概率会失败
func TestMayFail(t *testing.T) {
	var triggerCount = 5000
	initTest()
	httpServer := util.NewSimpleHttpServer(&util.SimpleHttpServerConf{
		FailRate:             0.1, //10%的概率失败
		ListeningPort:        SimpleWebServerPort,
		TriggerCount:         triggerCount,
		AllowDuplicateCalled: false,
	})
	go httpServer.Start()

	time.Sleep(3 * time.Second)
	//加一个任务
	if err := util.RegisterJob(SchedulerAddress, &model.Job{
		Name:                  "test-http-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Http",
		GlueSource: map[string]string{
			"method":     "POST",
			"url":        "http://localhost:" + strconv.Itoa(SimpleWebServerPort) + "/test",
			"timeout":    "30",
			"expectCode": "200",
			"expectBody": "OK",
			"debug":      "",
		},
	}); err != nil {
		panic(err)
	}

	triggers := make([]*model.Trigger, triggerCount)
	for i := 0; i < triggerCount; i++ {
		triggers[i] = &model.Trigger{
			Name:            "test-trigger-" + strconv.Itoa(i),
			JobID:           1,
			ScheduleType:    2,          //执行一次
			FailRetryCount:  100,        //失败几乎可以一直重试
			ExecuteTimeout:  2000000000, //2s
			TriggerNextTime: time.Now(),
			MisfireStrategy: constance.MisfireStrategyTypeDoNothing,
			Param: map[string]string{
				//把triggerID带过去，由SimpleHttpServer记录执行情况
				util.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}

	util.RegisterTriggers(SchedulerAddress, triggers)

	time.Sleep(15 * time.Second)
	httpServer.CheckResult(&util.CheckResultConf{
		AllSuccess:                    true,
		NoUncalledTriggers:            true,
		FailTriggerRateNotGreaterThan: 0.3,
	})
}
