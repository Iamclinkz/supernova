package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/stress-test/simple-http-server"
	"supernova/stress-test/util"
	"sync"
	"sync/atomic"
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

func initTest(schedulerCount, executorCount int, level klog.Level) {
	klog.SetLevel(level)
	util.StartHttpExecutors(util.GenExecutorInstanceConfWithCount(executorCount))
	time.Sleep(2 * time.Second)
	util.StartSchedulers(schedulerCount)
}

// TestWithoutFail 执行单次海量任务，任务默认成功，只有少数会失败（因为simple-http-server一瞬间建立太多连接，
// 可能会响应超时，从而http-executor因为超时，返回任务执行失败。
// 目标：可以看做是模拟正常情况下的高并发测试
func TestWithoutFail(t *testing.T) {
	start := time.Now()

	var triggerCount = 100000

	initTest(3, 3, klog.LevelWarn)
	httpServer := simple_http_server.NewSimpleHttpServer(
		&simple_http_server.SimpleHttpServerInitConf{
			FailRate:             0,
			ListeningPort:        SimpleWebServerPort,
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
				simple_http_server.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}

	util.RegisterTriggers(SchedulerAddress, triggers)

	log.Printf("register triggers successed, cost:%v\n", time.Since(start))
	httpServer.WaitResult(20 * time.Second)
}

// TestRandomFailJob 执行单次大量任务，任务有80%的概率第一次失败，模拟极端情况下，大量任务同时失败、超时需要重试的情况
// 目标：只要Scheduler和Executor都不宕机，那么任务一定能精准执行（成功）一次，即使失败多次。
func TestMayFail(t *testing.T) {
	var triggerCount = 10000
	initTest(3, 3, klog.LevelWarn)
	httpServer := simple_http_server.NewSimpleHttpServer(
		&simple_http_server.SimpleHttpServerInitConf{
			FailRate:              0.80, //80%的概率失败
			ListeningPort:         SimpleWebServerPort,
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
				simple_http_server.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}

	util.RegisterTriggers(SchedulerAddress, triggers)

	httpServer.WaitResult(20 * time.Second)
}

// TestForceKillScheduler 测试Scheduler宕机的情况。
// 目标：只要有其他可用的Scheduler实例，那么即使某个Scheduler死掉，任务可以正常执行，不会多触发、少触发
// 过程：使用SDK开启两个Scheduler，使用进程开启一个Scheduler，开启三个Executor。然后把Scheduler进程杀死，查看任务执行情况
func TestForceKillScheduler(t *testing.T) {
	const (
		BinPath       = "../scheduler/build/scheduler"
		HttpServePort = 7070
		LogLevel      = klog.LevelTrace
		TriggerCount  = 50000
	)

	var (
		err error
		pid atomic.Int32
	)

	//开2个Scheduler和3个Executor
	initTest(2, 3, klog.LevelWarn)

	httpServer := simple_http_server.NewSimpleHttpServer(&simple_http_server.SimpleHttpServerInitConf{
		FailRate:              0.10, //10%的概率失败
		ListeningPort:         SimpleWebServerPort,
		TriggerCount:          TriggerCount,
		AllowDuplicateCalled:  false,
		SuccessAfterFirstFail: true, //但是失败之后，重试一定成功
	}, &simple_http_server.SimpleHttpServerCheckConf{
		AllSuccess:         true,
		NoUncalledTriggers: true,
	})
	go httpServer.Start()

	//是否是手动删除
	manualKill := atomic.Bool{}

	schedulerStopWg := sync.WaitGroup{}
	schedulerStopWg.Add(1)

	go func() { //使用bin启动一个Scheduler，把日志输出到killed-scheduler.log中
		logFile, err := os.OpenFile("killed-scheduler.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		cmd := exec.Command(BinPath, fmt.Sprintf("-port=%d -logLevel=%d", HttpServePort, LogLevel))
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		err = cmd.Start()
		if err != nil {
			panic(err)
		}

		// 等待进程结束，并且判断是否是外围手动kill的。
		// Scheduler本身不会自动结束，所以如果外围没有手动kill且结束，只能说明启动出错，应该panic
		err = cmd.Wait()
		if !manualKill.Load() {
			panic("start error!")
		}
		pid.Store(int32(cmd.Process.Pid))
		klog.Warnf("scheduler bin was killed, error:%v", err)
		schedulerStopWg.Done()
	}()

	//等待SchedulerBin启动
	time.Sleep(1 * time.Second)

	if err = util.RegisterJob(SchedulerAddress, &model.Job{
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

	triggers := make([]*model.Trigger, TriggerCount)
	for i := 0; i < TriggerCount; i++ {
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
				simple_http_server.TriggerIDFieldName: strconv.Itoa(i),
			},
			FailRetryInterval: 0,
		}
	}
	util.RegisterTriggers(SchedulerAddress, triggers)

	time.Sleep(5 * time.Second)

	//杀死进程。这里先检查一手，别把init进程杀了。。。。。
	schedulerPid := pid.Load()
	err = util.KillProcessByPID(int(schedulerPid))
	if err != nil {
		panic(err)
	}

	httpServer.WaitResult(10 * time.Second)
}
