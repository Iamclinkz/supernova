package dao

import (
	"encoding/json"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"time"

	"gorm.io/gorm"
)

type Trigger struct {
	gorm.Model
	Name              string
	JobID             uint
	ScheduleType      constance.ScheduleType
	ScheduleConf      string
	MisfireStrategy   constance.MisfireStrategyType
	FailRetryCount    int
	ExecuteTimeout    time.Duration
	FailRetryInterval time.Duration
	TriggerLastTime   time.Time
	TriggerNextTime   time.Time
	Status            constance.TriggerStatus
	Param             string
	AtLeastOnce       bool
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
		Model: gorm.Model{
			ID:        mTrigger.ID,
			UpdatedAt: mTrigger.UpdatedAt,
		},
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
