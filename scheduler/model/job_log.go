package model

import "time"

// todo 不管成功失败，每次任务的执行搞一个JobLog
type JobLog struct {
	ID                     int       `gorm:"column:id;primaryKey;autoIncrement"`
	JobID                  int       `gorm:"column:job_id"`
	ExecutorAddress        string    `gorm:"column:executor_address"`
	ExecutorParam          string    `gorm:"column:executor_param"`
	ExecutorFailRetryCount int       `gorm:"column:executor_fail_retry_count"`
	TriggerTime            time.Time `gorm:"column:trigger_time"`
	TriggerCode            int       `gorm:"column:trigger_code"`
	TriggerMsg             string    `gorm:"column:trigger_msg"`
	HandleTime             time.Time `gorm:"column:handle_time"`
	HandleCode             int       `gorm:"column:handle_code"`
	HandleMsg              string    `gorm:"column:handle_msg"`
	AlarmStatus            int       `gorm:"column:alarm_status"`
}

// TableName 指定表名
func (j *JobLog) TableName() string {
	return "t_job_log"
}
