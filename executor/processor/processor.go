package processor

import (
	"supernova/pkg/api"
)

// PC 实在不知道起个啥名字了
type PC struct {
	Processor JobProcessor
	Config    *ProcessConfig
}

type ProcessConfig struct {
	Async          bool //是否异步执行
	MaxWorkerCount int  //同步和异步都要填，最大的工作go程数量
}

// JobProcessor 实现这个接口的类型可以借助executor sdk发布Executor服务
type JobProcessor interface {
	Process(job *api.Job) *api.JobResult
	GetGlueType() string
}
