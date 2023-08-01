package util

import (
	eapp "supernova/executor/app"
	sapp "supernova/scheduler/app"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type SupernovaTest struct {
	executors  []*eapp.Executor
	schedulers []*sapp.Scheduler
}

func StartTest(schedulerCount, executorCount int, level klog.Level, startFunc ExecutorStartFunc,
	extraConf any) *SupernovaTest {
	ret := new(SupernovaTest)

	klog.SetLevel(level)
	ret.executors = startFunc(GenExecutorInstanceConfWithCount(executorCount), extraConf)
	time.Sleep(2 * time.Second)
	ret.schedulers = StartSchedulers(schedulerCount)
	return ret
}

func (t *SupernovaTest) EndTest() {
	klog.Infof("start test end")

	wg := sync.WaitGroup{}
	wg.Add(len(t.executors))
	wg.Add(len(t.schedulers))

	for _, executor := range t.executors {
		go func(e *eapp.Executor) {
			e.Stop()
			wg.Done()
		}(executor)
	}

	for _, scheduler := range t.schedulers {
		go func(s *sapp.Scheduler) {
			s.Stop()
			wg.Done()
		}(scheduler)
	}

	wg.Wait()
	klog.Infof("SupernovaTest End")
}
