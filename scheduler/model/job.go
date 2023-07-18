package model

import (
	"errors"
	"supernova/pkg/api"
	"supernova/scheduler/constance"
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model `json:"-"`
	Name       string `gorm:"column:name;type:varchar(64);unique"`

	// 选择执行器的策略
	ExecutorRouteStrategy constance.ExecutorRouteStrategyType `gorm:"column:executor_route_strategy;type:tinyint(4);not null"`
	// 执行命令的对象（shell，python等）
	GlueType constance.GlueType `gorm:"column:glue_type;type:tinyint(4);not null"`
	// 命令源代码
	GlueSource string `gorm:"column:glue_source;type:MEDIUMTEXT;not null"`
	// 当前状态
	Status constance.JobStatus `gorm:"column:status;type:tinyint(4);not null"`
	// 执行的最大的执行时间，即Executor纯执行，最大时间
	ExecutorExecuteTimeout time.Duration `gorm:"column:executor_timeout;type:bigint;not null"`

	Tags string `gorm:"column:tags;not null"`
}

func (j *Job) TableName() string {
	return "t_job"
}

type RunJobRequest struct {
	// 执行命令的对象（shell，python等）
	GlueType constance.GlueType
	// 命令源代码
	GlueSource string
	// 执行的最大的执行时间，即Executor纯执行，最大时间
	ExecutorExecuteTimeout time.Duration
}

type RunJobResponse struct {
	Ok     bool
	Err    error
	Result string
}

func NewRunJobRequestFromJob(job *Job) *RunJobRequest {
	return &RunJobRequest{
		GlueType:               job.GlueType,
		GlueSource:             job.GlueSource,
		ExecutorExecuteTimeout: job.ExecutorExecuteTimeout,
	}
}

func (r *RunJobRequest) ToPb() *api.RunJobRequest {
	ret := new(api.RunJobRequest)
	ret.GlueType = int32(r.GlueType)
	ret.GlueSource = r.GlueSource
	ret.ExecutorExecuteTimeoutMs = r.ExecutorExecuteTimeout.Milliseconds()
	return ret
}

func (r *RunJobResponse) FromPb(resp *api.RunJobResponse) {
	r.Ok = resp.Ok
	r.Result = resp.Result
	if resp.Err != "" {
		r.Err = errors.New(resp.Err)
	}
}
