package processor

import (
	"supernova/pkg/api"
)

// 实现这个接口的类型可以借助executor sdk发布Executor服务
type JobProcessor interface {
	Process(job *api.Job) *api.JobResult
	GetGlueType() string
}
