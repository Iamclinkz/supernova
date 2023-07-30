package model

import (
	"supernova/scheduler/constance"
	"time"
)

type Job struct {
	UpdatedAt             time.Time
	ID                    uint
	Name                  string
	ExecutorRouteStrategy constance.ExecutorRouteStrategyType // 选择执行器的策略
	GlueType              string                              // 执行命令的对象（shell，python等）
	GlueSource            map[string]string                   // 命令源代码
	Status                constance.JobStatus                 // 当前状态
	Tags                  string
}
