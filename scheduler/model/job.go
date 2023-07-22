package model

import (
	"encoding/json"
	"supernova/scheduler/constance"

	"gorm.io/gorm"
)

type GlueSourceType map[string]string

type Job struct {
	gorm.Model `json:"-"`
	//todo 为了测试，允许name为null了，后期删掉
	//Name       string `gorm:"column:name;type:varchar(64);unique"`
	Name                  string                              `gorm:"column:name;type:varchar(64)"`
	ExecutorRouteStrategy constance.ExecutorRouteStrategyType `gorm:"column:executor_route_strategy;type:tinyint(4);not null"` // 选择执行器的策略
	GlueType              string                              `gorm:"column:glue_type;not null"`                               // 执行命令的对象（shell，python等）
	GlueSourceToDB        string                              `gorm:"column:glue_source"`
	GlueSource            map[string]string                   `gorm:"-"`                                      // 命令源代码
	Status                constance.JobStatus                 `gorm:"column:status;type:tinyint(4);not null"` // 当前状态
	Tags                  string                              `gorm:"column:tags;not null"`
}

func (j *Job) TableName() string {
	return "t_job"
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	return j.prepareGlueSource()
}

func (j *Job) BeforeUpdate(tx *gorm.DB) error {
	return j.prepareGlueSource()
}

func (j *Job) AfterFind(tx *gorm.DB) error {
	return j.parseGlueSource()
}

func (j *Job) prepareGlueSource() error {
	if j.GlueSource != nil && len(j.GlueSource) != 0 {
		jsonData, err := json.Marshal(j.GlueSource)
		if err != nil {
			return err
		}
		j.GlueSourceToDB = string(jsonData)
	}
	return nil
}

func (j *Job) parseGlueSource() error {
	if j.GlueSourceToDB != "" {
		var glueSourceMap map[string]string
		err := json.Unmarshal([]byte(j.GlueSourceToDB), &glueSourceMap)
		if err != nil {
			return err
		}
		j.GlueSource = glueSourceMap
	}
	return nil
}
