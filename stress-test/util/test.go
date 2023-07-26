package util

import (
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func InitTest(schedulerCount, executorCount int, level klog.Level) {
	klog.SetLevel(level)
	StartHttpExecutors(GenExecutorInstanceConfWithCount(executorCount))
	time.Sleep(2 * time.Second)
	StartSchedulers(schedulerCount)
}
