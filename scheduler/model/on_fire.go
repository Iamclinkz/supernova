package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"supernova/pkg/api"
	"supernova/scheduler/constance"
	"time"

	"gorm.io/gorm"
)

// OnFireLog 表示一次trigger的执行
type OnFireLog struct {
	//任务执行阶段取
	gorm.Model
	TriggerID         uint                   `gorm:"column:trigger_id;not null;index"`
	JobID             uint                   `gorm:"column:job_id;not null;index"`
	Status            constance.OnFireStatus `gorm:"column:status;type:tinyint(4);not null"`
	RetryCount        int                    `gorm:"column:retry_count;not null"`                  //当前重试次数
	LeftRetryCount    int                    `gorm:"column:left_retry_count"`                      //剩余的重试次数
	ExecutorInstance  string                 `gorm:"column:executor_instance"`                     //上一个执行的Executor的InstanceID
	RedoAt            time.Time              `gorm:"column:redo_at;type:timestamp;not null;index"` //超时时间
	ParamToDB         string                 `gorm:"column:param"`
	Param             map[string]string      `gorm:"-"`
	FailRetryInterval time.Duration          `gorm:"column:fail_retry_interval"` //失败重试间隔，为0则立刻重试

	//任务结束阶段使用
	Success bool   `gorm:"success"`
	Result  string `gorm:"result;type:varchar(256)"` //error or result

	//内部使用，不落库
	ShouldFireAt   time.Time
	ExecuteTimeout time.Duration
}

func (o *OnFireLog) TableName() string {
	return "t_on_fire"
}

func GenRunJobRequest(onFireLog *OnFireLog, job *Job) *api.RunJobRequest {
	return &api.RunJobRequest{
		OnFireLogID: uint32(onFireLog.ID),
		Job: &api.Job{
			GlueType:                 job.GlueType,
			Source:                   job.GlueSource,
			Param:                    onFireLog.Param,
			ExecutorExecuteTimeoutMs: onFireLog.ExecuteTimeout.Milliseconds(),
		},
	}
}

func (o *OnFireLog) BeforeCreate(tx *gorm.DB) error {
	return o.prepareParam()
}

func (o *OnFireLog) BeforeUpdate(tx *gorm.DB) error {
	return o.prepareParam()
}

func (o *OnFireLog) AfterFind(tx *gorm.DB) error {
	return o.parseParam()
}

func (o *OnFireLog) prepareParam() error {
	if o.Param != nil && len(o.Param) != 0 {
		jsonData, err := json.Marshal(o.Param)
		if err != nil {
			return err
		}
		o.ParamToDB = string(jsonData)
	}
	return nil
}

func (o *OnFireLog) parseParam() error {
	if o.ParamToDB != "" {
		var paramMap map[string]string
		err := json.Unmarshal([]byte(o.ParamToDB), &paramMap)
		if err != nil {
			return err
		}
		o.Param = paramMap
	}
	return nil
}

func OnFireLogsToString(onFireLogs []*OnFireLog) string {
	var builder strings.Builder
	for i, onFireLog := range onFireLogs {
		builder.WriteString(fmt.Sprintf("OnFireLog %d: %s\n", i+1, onFireLog.String()))
	}

	return builder.String()
}

func (o *OnFireLog) String() string {
	return fmt.Sprintf("OnFireLog(Model=%v, TriggerID=%d, JobID=%d, Status=%s, RetryCount=%d, LeftRetryCount=%d, ExecutorInstance=%s, RedoAt=%v, Param=%v, Success=%t, Result=%s, ShouldFireAt=%v, ExecuteTimeout=%v)",
		o.Model,
		o.TriggerID,
		o.JobID,
		o.Status,
		o.RetryCount,
		o.LeftRetryCount,
		o.ExecutorInstance,
		o.RedoAt,
		o.Param,
		o.Success,
		o.Result,
		o.ShouldFireAt,
		o.ExecuteTimeout,
	)
}
