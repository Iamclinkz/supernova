package model

import (
	"fmt"
	"strings"
	"supernova/pkg/api"
	"supernova/scheduler/constance"
	"time"

	"github.com/google/btree"
)

// OnFireLog 表示一次trigger的执行
type OnFireLog struct {
	//任务执行阶段取
	ID                uint
	UpdatedAt         time.Time
	TriggerID         uint
	JobID             uint
	Status            constance.OnFireStatus
	TryCount          int       //（失败）可执行次数
	LeftTryCount      int       //剩余的重试次数
	ExecutorInstance  string    //上一个执行的Executor的InstanceID
	RedoAt            time.Time //超时时间
	Param             map[string]string
	FailRetryInterval time.Duration //失败重试间隔，为0则立刻重试
	AtLeastOnce       bool          //语义，至少一次。如果为false，则为至多一次
	TraceContext      string        //trace相关

	Success bool
	Result  string //error or result

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
	return time.Now().Add(time.Duration(o.TryCount-o.LeftTryCount+1)*o.FailRetryInterval + o.ExecuteTimeout)
}

func (o *OnFireLog) Less(than btree.Item) bool {
	return o.RedoAt.Before(than.(*OnFireLog).RedoAt)
}
