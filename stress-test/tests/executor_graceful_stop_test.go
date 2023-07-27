package tests

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"supernova/pkg/conf"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	simple_http_server "supernova/stress-test/simple-http-server"
	"supernova/stress-test/util"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func TestExecutorGracefulStop(t *testing.T) {
	//测试使用
	const (
		BinPath                 = "../../executor-example/http-executor/build/http-executor"
		LogLevel                = klog.LevelInfo
		TriggerCount            = 50000
		MaxWaitGracefulStopTime = time.Second*5 + conf.SchedulerMaxCheckHealthDuration
	)

	//启动Executor使用
	const (
		ExecutorGrpcHost        = "9.134.5.191"
		ExecutorGrpcPort        = 20001
		ExecutorLogLevel        = klog.LevelInfo
		ExecutorHealthCheckPort = 8080
		ExecutorConsulHost      = "9.134.5.191"
		ExecutorConsulPort      = 8500
	)

	var (
		LogName = fmt.Sprintf("graceful-stop-executor-%v.log", time.Now().Format("15:04:05"))
		err     error
		pid     atomic.Int32
	)

	//开2个Scheduler和3个Executor
	util.InitTest(2, 2, LogLevel)

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

	//是否是手动触发退出
	manualQuit := atomic.Bool{}

	executorStopWg := sync.WaitGroup{}
	executorStopWg.Add(1)

	go func() {
		//使用bin启动一个Scheduler，把日志输出到graceful-stop-executor.log中
		logFile, err := os.OpenFile(LogName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		cmd := exec.Command(BinPath,
			fmt.Sprintf("-grpcHost=%v", ExecutorGrpcHost), fmt.Sprintf("-grpcPort=%v", ExecutorGrpcPort),
			fmt.Sprintf("-logLevel=%v", ExecutorLogLevel), fmt.Sprintf("-healthCheckPort=%v", ExecutorHealthCheckPort),
			fmt.Sprintf("-consulHost=%v", ExecutorConsulHost), fmt.Sprintf("-consulPort=%v", ExecutorConsulPort))
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
		if !manualQuit.Load() {
			panic("start error!")
		}

		klog.Infof("executor stopped, reason:%v", err)
		executorStopWg.Done()
	}()

	//等待ExecutorBin启动
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

		//发送信号。这里先检查一手，别把init进程杀了。。。。。
		executorPid := pid.Load()
		if executorPid == 0 {
			panic("")
		}
		manualQuit.Store(true)
		err = util.SendSignalByPid(int(executorPid), syscall.SIGTERM)
		if err != nil {
			panic(err)
		}
	}()

	util.RegisterTriggers(util.SchedulerAddress, triggers)
	httpServer.WaitResult(MaxWaitGracefulStopTime, false)
}
