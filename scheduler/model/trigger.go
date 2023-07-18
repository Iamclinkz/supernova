package model

import (
	"fmt"
	"strings"
	"supernova/scheduler/constance"
	"supernova/scheduler/util"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Trigger struct {
	gorm.Model `json:"-"`
	//Name       string `gorm:"column:name;type:varchar(64);unique"`
	Name string `gorm:"column:name;type:varchar(64)"`
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
	//JobUpdateTime time.Time `json:"-"`
}

func (t *Trigger) String() string {
	return fmt.Sprintf("Trigger{ID: %d, Name: %s, JobID: %d, ScheduleType: %d, ScheduleConf: %s,"+
		" MisfireStrategy: %s, FailRetryCount: %d, TriggerTimeout: %s, TriggerLastTime: %s, TriggerNextTime: %s, Status: %s}",
		t.ID,
		t.Name,
		t.JobID,
		t.ScheduleType,
		t.ScheduleConf,
		t.MisfireStrategy,
		t.FailRetryCount,
		t.TriggerTimeout,
		t.TriggerLastTime.Format(time.RFC3339),
		t.TriggerNextTime.Format(time.RFC3339),
		t.Status,
	)
}

func (t *Trigger) TableName() string {
	return "t_trigger"
}

// SecondParser 精确到秒的parser
var SecondParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Second)

// NextExecutionTime 获取下一次Fire的时间
func (t *Trigger) NextExecutionTime() (time.Time, error) {
	switch t.ScheduleType {
	case constance.ScheduleTypeNone, constance.ScheduleTypeOnce:
		//如果是不执行，或者只执行一次，那么下次执行时间设置为很久以后
		return util.VeryLateTime(), nil
	case constance.ScheduleTypeCron:
		cronParser, err := SecondParser.Parse(t.ScheduleConf)
		if err != nil {
			return util.VeryLateTime(), err
		}
		return cronParser.Next(t.TriggerNextTime), nil
	default:
		return util.VeryLateTime(), fmt.Errorf("unknown schedule type: %v", t.ScheduleType)
	}
}

// OnFire Trigger准备被fire时调用，目前只是更新下次fire时间，并检查本次fire时间是否正确
func (t *Trigger) OnFire() error {
	var err error
	t.TriggerLastTime = t.TriggerNextTime
	t.TriggerNextTime, err = t.NextExecutionTime()
	return err
}

func TriggersToString(triggers []*Trigger) string {
	var builder strings.Builder
	builder.WriteString("find triggers:\n")
	for i, trigger := range triggers {
		builder.WriteString(fmt.Sprintf("Trigger %d: %s\n", i+1, trigger.String()))
	}

	return builder.String()
}
