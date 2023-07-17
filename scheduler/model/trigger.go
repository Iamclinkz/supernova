package model

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"supernova/pkg/constance"
	"supernova/scheduler/util"
	"time"
)

type Trigger struct {
	gorm.Model
	Name string `gorm:"column:name;type:varchar(64);unique"`

	//关联的任务ID
	JobID uint `gorm:"not null;index"`
	// 选择调度的策略（None表示不触发，Cron表示使用Cron表达式，
	ScheduleType constance.ScheduleType `gorm:"column:schedule_type;type:tinyint(4);not null"`
	// 配合调度策略，这里的配置信息
	ScheduleConf string `gorm:"column:schedule_conf;type:varchar(128)"`
	// 如果错过了本次执行，应该采取的策略。和Fail-Retry概念不一样，misfire指任务本身没有发生错误，但是
	// 例如Scheduler有太多任务做不过来等原因，使得调度到本trigger的时候，已经错过了本trigger的执行时间。
	//todo 待实现
	MisfireStrategy constance.MisfireStrategyType `gorm:"column:misfire_strategy;type:tinyint(4);not null"`
	// 如果出错，最大重试次数
	FailRetryCount int `gorm:"column:executor_fail_retry_count;not null"`
	// 在t_on_fire表中呆多长时间，算超时
	TriggerTimeout time.Duration `gorm:"column:executor_timeout;type:bigint;not null"`
	// 上次触发时间
	TriggerLastTime time.Time `gorm:"column:trigger_last_time;type:timestamp;not null"`
	// 下次触发时间
	TriggerNextTime time.Time `gorm:"column:trigger_next_time;type:timestamp;not null;index"`
	// 当前状态
	Status constance.TriggerStatus `gorm:"column:status;type:tinyint(4);not null"`
	//关联的Job的修改时间
	JobUpdateTime time.Time
}

func (t *Trigger) TableName() string {
	return "t_trigger"
}

func (t *Trigger) NextExecutionTime() (time.Time, error) {
	switch t.ScheduleType {
	case constance.ScheduleTypeNone, constance.ScheduleTypeOnce:
		//如果是不执行，或者只执行依次，那么下次执行时间设置为很久以后
		return util.VeryLongTime(), nil
	case constance.ScheduleTypeCron:
		cronParser, err := cron.ParseStandard(t.ScheduleConf)
		if err != nil {
			return util.VeryLongTime(), err
		}
		return cronParser.Next(time.Now()), nil
	default:
		return util.VeryLongTime(), fmt.Errorf("unknown schedule type: %v", t.ScheduleType)
	}
}

func (t *Trigger) OnFire() error {
	var err error
	t.TriggerLastTime = time.Now()
	t.TriggerNextTime, err = t.NextExecutionTime()
	return err
}
