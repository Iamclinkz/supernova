package main

import (
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/stress-test/util"
	"testing"
	"time"

	"log"

	"github.com/cloudwego/kitex/pkg/klog"
)

var (
	SimpleWebServerPort = 9000
	SchedulerPort       = util.SchedulerServePortStart + 1
	SchedulerAddress    = "http://localhost:" + strconv.Itoa(SchedulerPort)
)

func initTest() {
	klog.SetLevel(klog.LevelWarn)
	util.StartHttpExecutors(util.GenExecutorInstanceConfWithCount(5))
	time.Sleep(2 * time.Second)
	util.StartSchedulers(5)
}

// TestWithoutFail 执行单次，任务不会失败
func TestWithoutFail(t *testing.T) {
	start := time.Now()

	var triggerCount = 100000

	initTest()
	httpServer := util.NewSimpleHttpServer(&util.SimpleHttpServerConf{
		FailRate:             0,
		ListeningPort:        SimpleWebServerPort,
		TriggerCount:         triggerCount,
		AllowDuplicateCalled: false,
	})
	go httpServer.Start()

	time.Sleep(1 * time.Second)
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

	time.Sleep(20 * time.Second)

	httpServer.CheckResult(&util.CheckResultConf{
		AllSuccess:                    true,
		NoUncalledTriggers:            true,
		FailTriggerRateNotGreaterThan: 0.1,
	})
}

// TestRandomFailJob 执行单次，任务有概率会失败
func TestMayFail(t *testing.T) {
	var triggerCount = 10000
	initTest()
	httpServer := util.NewSimpleHttpServer(&util.SimpleHttpServerConf{
		FailRate:              0.80, //80%的概率失败
		ListeningPort:         SimpleWebServerPort,
		TriggerCount:          triggerCount,
		AllowDuplicateCalled:  false,
		SuccessAfterFirstFail: true, //但是失败之后，重试一定成功
	})
	go httpServer.Start()

	time.Sleep(1 * time.Second)
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

	time.Sleep(20 * time.Second)
	httpServer.CheckResult(&util.CheckResultConf{
		AllSuccess:                    false,
		NoUncalledTriggers:            true,
		FailTriggerRateNotGreaterThan: 0.01,
	})
}
