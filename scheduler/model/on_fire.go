package model

import (
	"gorm.io/gorm"
	"supernova/pkg/api"
	"supernova/scheduler/constance"
	"time"
)

// OnFireLog 表示一次trigger的执行
type OnFireLog struct {
	//任务执行阶段取
	gorm.Model
	TriggerID        uint                   `gorm:"column:trigger_id;not null;index"`
	JobID            uint                   `gorm:"column:job_id;not null;index"`
	Status           constance.OnFireStatus `gorm:"column:status;type:tinyint(4);not null"`
	RetryCount       int                    `gorm:"column:retry_count;not null"`                  //当前重试次数
	LeftRetryCount   int                    `gorm:"column:left_retry_count"`                      //剩余的重试次数
	ExecutorInstance string                 `gorm:"column:executor_instance"`                     //上一个执行的Executor的InstanceID
	RedoAt           time.Time              `gorm:"column:redo_at;type:timestamp;not null;index"` //超时时间
	Param            map[string]string      `gorm:"type:json"`                                    //trigger参数

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
