package dao

import (
	"encoding/json"
	"gorm.io/gorm"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
)

type GlueSourceType map[string]string

type Job struct {
	gorm.Model            `json:"-"`
	Name                  string                              `gorm:"column:name;type:varchar(64)"` //todo 为了测试，允许name为null了，后期删掉
	ExecutorRouteStrategy constance.ExecutorRouteStrategyType `gorm:"column:executor_route_strategy;type:tinyint(4);not null"`
	GlueType              string                              `gorm:"column:glue_type;not null"`
	GlueSource            string                              `gorm:"column:glue_source"`
	Status                constance.JobStatus                 `gorm:"column:status;type:tinyint(4);not null"`
	Tags                  string                              `gorm:"column:tags;not null"`
}

func (j *Job) TableName() string {
	return "t_job"
}

func FromModelJob(mJob *model.Job) (*Job, error) {
	var glueSourceStr string
	if mJob.GlueSource != nil && len(mJob.GlueSource) > 0 {
		glueSource, err := json.Marshal(mJob.GlueSource)
		if err != nil {
			return nil, err
		}
		glueSourceStr = string(glueSource)
	}

	return &Job{
		Model: gorm.Model{
			ID:        mJob.ID,
			UpdatedAt: mJob.UpdatedAt,
		},
		Name:                  mJob.Name,
		ExecutorRouteStrategy: mJob.ExecutorRouteStrategy,
		GlueType:              mJob.GlueType,
		GlueSource:            glueSourceStr,
		Status:                mJob.Status,
		Tags:                  mJob.Tags,
	}, nil
}

func ToModelJob(dJob *Job) (*model.Job, error) {
	var glueSource map[string]string
	err := json.Unmarshal([]byte(dJob.GlueSource), &glueSource)
	if err != nil {
		return nil, err
	}

	return &model.Job{
		ID:                    dJob.ID,
		UpdatedAt:             dJob.UpdatedAt,
		Name:                  dJob.Name,
		ExecutorRouteStrategy: dJob.ExecutorRouteStrategy,
		GlueType:              dJob.GlueType,
		GlueSource:            glueSource,
		Status:                dJob.Status,
		Tags:                  dJob.Tags,
	}, nil
}
