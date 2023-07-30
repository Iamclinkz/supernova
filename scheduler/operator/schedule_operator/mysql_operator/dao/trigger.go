package dao

import (
	"encoding/json"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"time"
)

type Trigger struct {
	ID                uint                          `gorm:"column:id;primarykey"`
	UpdatedAt         time.Time                     `gorm:"column:updated_at"`
	Name              string                        `gorm:"column:name;type:varchar(64)"`
	JobID             uint                          `gorm:"not null;index"`
	ScheduleType      constance.ScheduleType        `gorm:"column:schedule_type;type:tinyint(4);not null"`
	ScheduleConf      string                        `gorm:"column:schedule_conf;type:varchar(128)"`
	MisfireStrategy   constance.MisfireStrategyType `gorm:"column:misfire_strategy;type:tinyint(4);not null"`
	FailRetryCount    int                           `gorm:"column:executor_fail_retry_count;not null"`
	ExecuteTimeout    time.Duration                 `gorm:"column:execute_timeout;type:bigint;not null"`
	FailRetryInterval time.Duration                 `gorm:"column:fail_retry_interval"`
	TriggerLastTime   time.Time                     `gorm:"column:trigger_last_time;type:timestamp;not null"`
	TriggerNextTime   time.Time                     `gorm:"column:trigger_next_time;type:timestamp;not null;index"`
	Status            constance.TriggerStatus       `gorm:"column:status;type:tinyint(4);not null"`
	Param             string                        `gorm:"column:param"`
	AtLeastOnce       bool                          `gorm:"column:at_least_once"`
}

func (t *Trigger) TableName() string {
	return "t_trigger"
}

func FromModelTrigger(mTrigger *model.Trigger) (*Trigger, error) {
	param, err := json.Marshal(mTrigger.Param)
	if err != nil {
		return nil, err
	}

	return &Trigger{
		ID:                mTrigger.ID,
		UpdatedAt:         mTrigger.UpdatedAt,
		Name:              mTrigger.Name,
		JobID:             mTrigger.JobID,
		ScheduleType:      mTrigger.ScheduleType,
		ScheduleConf:      mTrigger.ScheduleConf,
		MisfireStrategy:   mTrigger.MisfireStrategy,
		FailRetryCount:    mTrigger.FailRetryCount,
		ExecuteTimeout:    mTrigger.ExecuteTimeout,
		FailRetryInterval: mTrigger.FailRetryInterval,
		TriggerLastTime:   mTrigger.TriggerLastTime,
		TriggerNextTime:   mTrigger.TriggerNextTime,
		Status:            mTrigger.Status,
		Param:             string(param),
		AtLeastOnce:       mTrigger.AtLeastOnce,
	}, nil
}

func ToModelTrigger(dTrigger *Trigger) (*model.Trigger, error) {
	var param map[string]string
	err := json.Unmarshal([]byte(dTrigger.Param), &param)
	if err != nil {
		return nil, err
	}

	return &model.Trigger{
		ID:                dTrigger.ID,
		UpdatedAt:         dTrigger.UpdatedAt,
		Name:              dTrigger.Name,
		JobID:             dTrigger.JobID,
		ScheduleType:      dTrigger.ScheduleType,
		ScheduleConf:      dTrigger.ScheduleConf,
		MisfireStrategy:   dTrigger.MisfireStrategy,
		FailRetryCount:    dTrigger.FailRetryCount,
		ExecuteTimeout:    dTrigger.ExecuteTimeout,
		FailRetryInterval: dTrigger.FailRetryInterval,
		TriggerLastTime:   dTrigger.TriggerLastTime,
		TriggerNextTime:   dTrigger.TriggerNextTime,
		Status:            dTrigger.Status,
		Param:             param,
		AtLeastOnce:       dTrigger.AtLeastOnce,
	}, nil
}
