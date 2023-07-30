package model

import (
	"fmt"
	"github.com/google/btree"
	"strings"
	"supernova/pkg/api"
	"supernova/scheduler/constance"
	"time"
)

// OnFireLog 表示一次trigger的执行
type OnFireLog struct {
	//任务执行阶段取
	ID                uint
	UpdatedAt         time.Time
	TriggerID         uint                   `gorm:"column:trigger_id;not null;"`
	JobID             uint                   `gorm:"column:job_id;not null;"`
	Status            constance.OnFireStatus `gorm:"column:status;type:tinyint(4);not null"`
	TryCount          int                    `gorm:"column:try_count;not null"`     //（失败）可执行次数
	LeftTryCount      int                    `gorm:"column:left_try_count"`         //剩余的重试次数
	ExecutorInstance  string                 `gorm:"column:executor_instance"`      //上一个执行的Executor的InstanceID
	RedoAt            time.Time              `gorm:"column:redo_at;not null;index"` //超时时间
	Param             map[string]string      `gorm:"-"`
	FailRetryInterval time.Duration          `gorm:"column:fail_retry_interval"` //失败重试间隔，为0则立刻重试
	AtLeastOnce       bool                   `gorm:"column:at_least_once"`       //语义，至少一次。如果为false，则为至多一次
	TraceContext      string                 `gorm:"trace_context"`              //trace相关

	//任务结束阶段使用
	Success bool   `gorm:"success"`
	Result  string `gorm:"result;type:varchar(256)"` //error or result

	//内部使用，不落库
	ShouldFireAt   time.Time
	ExecuteTimeout time.Duration
}

func GenRunJobRequest(onFireLog *OnFireLog, job *Job, traceContext map[string]string) *api.RunJobRequest {
	return &api.RunJobRequest{
		OnFireLogID:  uint32(onFireLog.ID),
		TraceContext: traceContext,
		Job: &api.Job{
			GlueType:                 job.GlueType,
			Source:                   job.GlueSource,
			Param:                    onFireLog.Param,
			ExecutorExecuteTimeoutMs: onFireLog.ExecuteTimeout.Milliseconds(),
		},
	}
}

func OnFireLogsToString(onFireLogs []*OnFireLog) string {
	var builder strings.Builder
	for i, onFireLog := range onFireLogs {
		builder.WriteString(fmt.Sprintf("OnFireLog %d: %s\n", i+1, onFireLog.String()))
	}

	return builder.String()
}

func (o *OnFireLog) String() string {
	return fmt.Sprintf("OnFireLog(ID=%v, UpdatedAt=%v, TriggerID=%d, JobID=%d, Status=%s, TryCount=%d, LeftTryCount=%d, ExecutorInstance=%s, RedoAt=%v, Param=%v, Success=%t, Result=%s, ShouldFireAt=%v, ExecuteTimeout=%v)",
		o.ID,
		o.UpdatedAt,
		o.TriggerID,
		o.JobID,
		o.Status,
		o.TryCount,
		o.LeftTryCount,
		o.ExecutorInstance,
		o.RedoAt,
		o.Param,
		o.Success,
		o.Result,
		o.ShouldFireAt,
		o.ExecuteTimeout,
	)
}

// GetNextRedoAt 计算下一次应该执行的时间（即本次的超时时间）。计算逻辑：
// 旧超时时间 +  （经过了几次超时+1） * 用户指定的重试间隔 + 用户指定的Task最大执行时间
func (o *OnFireLog) GetNextRedoAt() time.Time {
	return o.RedoAt.Add(time.Duration(o.TryCount-o.LeftTryCount+1)*o.FailRetryInterval + o.ExecuteTimeout)
}

func (o *OnFireLog) Less(than btree.Item) bool {
	return o.RedoAt.Before(than.(*OnFireLog).RedoAt)
}
