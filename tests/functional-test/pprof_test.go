package functional_test

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	simple_http_server "supernova/tests/simple-http-server"
	"supernova/tests/util"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func TestPprof(t *testing.T) {
	start := time.Now()

	var triggerCount = 100000
	go func() {
		err := http.ListenAndServe("0.0.0.0:6060", nil)
		if err != nil {
			panic(err)
		}
	}()
	supernovaTest := util.StartTest(2, 3, klog.LevelError, util.StartHttpExecutors, nil)
	defer supernovaTest.EndTest()

	httpServer := simple_http_server.NewSimpleHttpServer(
		&simple_http_server.SimpleHttpServerInitConf{
			FailRate:             0,
			ListeningPort:        util.SimpleWebServerPort,
			TriggerCount:         triggerCount,
			AllowDuplicateCalled: false,
		},
		&simple_http_server.SimpleHttpServerCheckConf{
			AllSuccess:                    true,
			NoUncalledTriggers:            true,
			FailTriggerRateNotGreaterThan: 1,
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
			ScheduleType:    2,          //执行一次
			FailRetryCount:  5,          //失败重试五次。因为simple_http_service每次都需要开go程执行请求，瞬间很多个请求打过去可能造成失败的情况。。
			ExecuteTimeout:  3000000000, //3s
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

	log.Printf("register triggers successed, cost:%v\n", time.Since(start))
	time.Sleep(1000 * time.Minute)
}
