package tests

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

	"github.com/cloudwego/kitex/pkg/klog"
)

// TestForceKillScheduler 测试Scheduler宕机的情况。
// 目标：只要有其他可用的Scheduler实例，那么即使某个Scheduler死掉，任务可以正常执行，不会多触发、少触发
// 过程：使用SDK开启两个Scheduler，使用进程开启一个Scheduler，开启三个Executor。然后把Scheduler进程杀死，查看任务执行情况
func TestForceKillScheduler(t *testing.T) {
	const (
		BinPath       = "../../scheduler/build/scheduler"
		HttpServePort = 7070
		LogLevel      = klog.LevelError
		TriggerCount  = 50000
	)

	var (
		err error
		pid atomic.Int32
	)

	//开2个Scheduler和3个Executor
	util.InitTest(2, 3, LogLevel)

	httpServer := simple_http_server.NewSimpleHttpServer(&simple_http_server.SimpleHttpServerInitConf{
		FailRate:              0.10, //10%的概率失败
		ListeningPort:         util.SimpleWebServerPort,
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
		cmd := exec.Command(BinPath,
			fmt.Sprintf("-port=%d", HttpServePort), fmt.Sprintf("-logLevel=%d", klog.LevelTrace))
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		//存储一下PID
		pid.Store(int32(cmd.Process.Pid))

		// 等待进程结束，并且判断是否是外围手动kill的。
		// Scheduler本身不会自动结束，所以如果外围没有手动kill且结束，只能说明启动出错，应该panic
		err = cmd.Wait()
		if !manualKill.Load() {
			panic("start error!")
		}

		klog.Error("scheduler was killed, reason:%v", err)
		schedulerStopWg.Done()
	}()

	//等待SchedulerBin启动
	time.Sleep(1 * time.Second)

	if err = util.RegisterJob(util.SchedulerAddress, &model.Job{
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

	go func() {
		time.Sleep(5 * time.Second)

		//杀死进程。这里先检查一手，别把init进程杀了。。。。。
		schedulerPid := pid.Load()
		if schedulerPid == 0 {
			panic("")
		}
		manualKill.Store(true)
		err = util.KillProcessByPID(int(schedulerPid))
		if err != nil {
			panic(err)
		}
	}()

	util.RegisterTriggers(util.SchedulerAddress, triggers)
	httpServer.WaitResult(10 * time.Second)
}