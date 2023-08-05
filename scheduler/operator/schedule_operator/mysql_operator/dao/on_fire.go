package dao

import (
	"encoding/json"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"time"
)

type OnFireLog struct {
	ID                uint                   `gorm:"column:id;primarykey"`
	UpdatedAt         time.Time              `gorm:"column:updated_at"`
	TriggerID         uint                   `gorm:"column:trigger_id;"`
	JobID             uint                   `gorm:"column:job_id;"`
	Status            constance.OnFireStatus `gorm:"column:status;type:tinyint(4);"`
	TryCount          int                    `gorm:"column:try_count;"`
	LeftTryCount      int                    `gorm:"column:left_try_count"`
	ExecutorInstance  string                 `gorm:"column:executor_instance"`
	RedoAt            time.Time              `gorm:"column:redo_at;not null;index"`
	Param             string                 `gorm:"column:param"`
	FailRetryInterval time.Duration          `gorm:"column:fail_retry_interval"`
	AtLeastOnce       bool                   `gorm:"column:at_least_once"`
	TraceContext      string                 `gorm:"trace_context"`
	Success           bool                   `gorm:"success"`
	Result            string                 `gorm:"result;type:varchar(256)"`
	ExecuteTimeout    time.Duration          `gorm:"execute_timeout"`
}

func (o *OnFireLog) TableName() string {
	return "t_on_fire"
}

func FromModelOnFireLog(mOnFireLog *model.OnFireLog) (*OnFireLog, error) {
	var paramStr string
	if mOnFireLog.Param != nil && len(mOnFireLog.Param) > 0 {
		param, err := json.Marshal(mOnFireLog.Param)
		if err != nil {
			return nil, err
		}
		paramStr = string(param)
	}

	return &OnFireLog{
		ID:                mOnFireLog.ID,
		UpdatedAt:         mOnFireLog.UpdatedAt,
		TriggerID:         mOnFireLog.TriggerID,
		JobID:             mOnFireLog.JobID,
		Status:            mOnFireLog.Status,
		TryCount:          mOnFireLog.TryCount,
		LeftTryCount:      mOnFireLog.LeftTryCount,
		ExecutorInstance:  mOnFireLog.ExecutorInstance,
		RedoAt:            mOnFireLog.RedoAt,
		Param:             paramStr,
		FailRetryInterval: mOnFireLog.FailRetryInterval,
		AtLeastOnce:       mOnFireLog.AtLeastOnce,
		TraceContext:      mOnFireLog.TraceContext,
		Success:           mOnFireLog.Success,
		Result:            mOnFireLog.Result,
		ExecuteTimeout:    mOnFireLog.ExecuteTimeout,
	}, nil
}

func ToModelOnFireLog(dOnFireLog *OnFireLog) (*model.OnFireLog, error) {
	var param map[string]string
	if len(dOnFireLog.Param) != 0 {
		err := json.Unmarshal([]byte(dOnFireLog.Param), &param)
		if err != nil {
			return nil, err
		}
	}

	return &model.OnFireLog{
		ID:                dOnFireLog.ID,
		UpdatedAt:         dOnFireLog.UpdatedAt,
		TriggerID:         dOnFireLog.TriggerID,
		JobID:             dOnFireLog.JobID,
		Status:            dOnFireLog.Status,
		TryCount:          dOnFireLog.TryCount,
		LeftTryCount:      dOnFireLog.LeftTryCount,
		ExecutorInstance:  dOnFireLog.ExecutorInstance,
		RedoAt:            dOnFireLog.RedoAt,
		Param:             param,
		FailRetryInterval: dOnFireLog.FailRetryInterval,
		AtLeastOnce:       dOnFireLog.AtLeastOnce,
		TraceContext:      dOnFireLog.TraceContext,
		Success:           dOnFireLog.Success,
		Result:            dOnFireLog.Result,
		ExecuteTimeout:    dOnFireLog.ExecuteTimeout,
	}, nil
}
