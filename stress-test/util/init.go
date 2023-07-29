package util

import (
	eapp "supernova/executor/app"
	sapp "supernova/scheduler/app"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type SupernovaTest struct {
	executors  []*eapp.Executor
	schedulers []*sapp.Scheduler
}

func StartTest(schedulerCount, executorCount int, level klog.Level) *SupernovaTest {
	ret := new(SupernovaTest)

	klog.SetLevel(level)
	ret.executors = StartHttpExecutors(GenExecutorInstanceConfWithCount(executorCount))
	time.Sleep(2 * time.Second)
	ret.schedulers = StartSchedulers(schedulerCount)
	return ret
}

func (t *SupernovaTest) EndTest() {
	for _, executor := range t.executors {
		executor.Stop()
	}

	for _, scheduler := range t.schedulers {
		scheduler.Stop()
	}

	klog.Infof("SupernovaTest End")
}
