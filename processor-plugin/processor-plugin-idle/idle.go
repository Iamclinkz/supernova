package processor_plugin_idle

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"math/rand"
	"supernova/executor/processor"
	"supernova/pkg/api"
	"time"
)

type IdleProcessorConfig struct {
	DoLog    bool
	DoSleep  bool          //是否睡眠
	SleepMin time.Duration //最大睡眠
	SleepMax time.Duration //最小睡眠
	DoFail   bool
	FailRate float32
}

// IdleProcessor 啥也不干的一个Processor的实现，使用sleep一段时间，模仿任务执行，随机返回成功或者失败
type IdleProcessor struct {
	config *IdleProcessorConfig
}

func NewIdle(config *IdleProcessorConfig) *IdleProcessor {
	if config.DoSleep && config.SleepMin > config.SleepMax {
		panic("")
	}

	return &IdleProcessor{
		config: config,
	}
}

var _ processor.JobProcessor = (*IdleProcessor)(nil)

// Process
// "glueType": "Idle",
//
//	"glueSource": {
//	    "message": "Hello",
//	}
func (s *IdleProcessor) Process(job *api.Job) *api.JobResult {
	msg, ok := job.Source["message"]
	if !ok {
		panic("")
	}

	if s.config.DoLog {
		klog.Debugf("idle processor receive:%v", msg)
	}

	if s.config.DoSleep {
		time.Sleep(time.Duration(rand.Int63n(int64(s.config.SleepMax-s.config.SleepMin)) + int64(s.config.SleepMin)))
	}

	result := new(api.JobResult)
	if s.config.DoFail && rand.Float32() < s.config.FailRate {
		//失败
		result.Ok = false
		result.Err = "random fail"
	} else {
		result.Ok = true
		result.Result = "just ok"
	}

	return result
}

func (s *IdleProcessor) GetGlueType() string {
	return "Idle"
}
