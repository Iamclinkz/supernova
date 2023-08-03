package functional_test

import (
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	simple_http_server "supernova/tests/simple-http-server"
	"supernova/tests/util"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// TestFail 执行单次大量任务，任务有80%的概率第一次失败，模拟极端情况下，大量任务同时失败、超时需要重试的情况
// 目标：只要Scheduler和Executor都不宕机，那么任务一定能精准执行（成功）一次，即使失败多次。
func TestFail(t *testing.T) {
	var triggerCount = 10000
	supernovaTest := util.StartTest(2, 2, klog.LevelDebug, util.StartHttpExecutors, nil)
	defer supernovaTest.EndTest()

	httpServer := simple_http_server.NewSimpleHttpServer(
		&simple_http_server.SimpleHttpServerInitConf{
			FailRate:              0.80, //80%的概率失败
			ListeningPort:         util.SimpleWebServerPort,
			TriggerCount:          triggerCount,
			AllowDuplicateCalled:  false,
			SuccessAfterFirstFail: true, //但是失败之后，重试一定成功
		}, &simple_http_server.SimpleHttpServerCheckConf{
			AllSuccess:                    false,
			NoUncalledTriggers:            true,
			FailTriggerRateNotGreaterThan: 0.01,
		})
	go httpServer.Start()

	time.Sleep(1 * time.Second)
	//加一个任务
	if err := util.RegisterJob(util.SchedulerAddress, &model.Job{
		Name:                  "test-http-job",
		ExecutorRouteStrategy: constance.ExecutorRouteStrategyTypeRandom, //随机路由
		GlueType:              "Http",
		GlueSource: map[string]string{
			"method":     "POST",
			"url":        "http://localhost:" + strconv.Itoa(util.SimpleWebServerPort) + "/test",
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
			ScheduleType:    2,               //执行一次
			FailRetryCount:  100,             //失败几乎可以一直重试
			ExecuteTimeout:  5 * time.Second, //2s
			TriggerNextTime: time.Now(),
			MisfireStrategy: constance.MisfireStrategyTypeDoNothing,
			Param: map[string]string{
				//把triggerID带过去，由SimpleHttpServer记录执行情况
				simple_http_server.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}

	util.RegisterTriggers(util.SchedulerAddress, triggers)

	httpServer.WaitResult(20*time.Second, true)
}
