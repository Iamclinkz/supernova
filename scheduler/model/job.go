package model

import (
	"gorm.io/gorm"
	"supernova/scheduler/constance"
)

type Job struct {
	gorm.Model `json:"-"`
	//todo 为了测试，允许name为null了，后期删掉
	//Name       string `gorm:"column:name;type:varchar(64);unique"`
	Name                  string                              `gorm:"column:name;type:varchar(64)"`
	ExecutorRouteStrategy constance.ExecutorRouteStrategyType `gorm:"column:executor_route_strategy;type:tinyint(4);not null"` // 选择执行器的策略
	GlueType              string                              `gorm:"column:glue_type;not null"`                               // 执行命令的对象（shell，python等）
	GlueSource            map[string]string                   `gorm:"column:glue_source;type:json;not null"`                   // 命令源代码
	Status                constance.JobStatus                 `gorm:"column:status;type:tinyint(4);not null"`                  // 当前状态
	Tags                  string                              `gorm:"column:tags;not null"`
}

func (j *Job) TableName() string {
	return "t_job"
}
