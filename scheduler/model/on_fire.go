package model

import (
	"gorm.io/gorm"
	"supernova/scheduler/constance"
	"time"
)

// OnFireLog 表示一次trigger的执行
type OnFireLog struct {
	gorm.Model
	TriggerID        uint                   `gorm:"column:trigger_id;not null"`
	JobID            uint                   `gorm:"column:job_id;not null"`
	Status           constance.OnFireStatus `gorm:"column:status;type:tinyint(4);not null"`
	RetryCount       int                    `gorm:"column:retry_count;not null"`
	ExecutorInstance string                 `gorm:"column:executor_instance"`
	TimeoutAt        time.Time              `gorm:"column:timeout_at;type:timestamp;not null;index"`
	ShouldFireAt     time.Time              `gorm:"column:should_fire_at;type:timestamp;not null"` //todo 这个字段应该填一下
}

func (o *OnFireLog) TableName() string {
	return "t_on_fire"
}
