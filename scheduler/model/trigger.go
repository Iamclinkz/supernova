package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
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
	// 如果出错，最大重试次数。
	FailRetryCount int `gorm:"column:executor_fail_retry_count;not null"`
	// executor执行的最大时间。要求用户一定指定。如果超时，则减少OnFireLog中的RetryCount字段
	ExecuteTimeout    time.Duration `gorm:"column:execute_timeout;type:bigint;not null"`
	FailRetryInterval time.Duration `gorm:"column:fail_retry_interval"` //失败重试间隔，为0则立刻重试
	// 上次触发时间
	TriggerLastTime time.Time `gorm:"column:trigger_last_time;type:timestamp;not null"`
	// 下次触发时间
	TriggerNextTime time.Time `gorm:"column:trigger_next_time;type:timestamp;not null;index"`
	// 当前状态
	Status    constance.TriggerStatus `gorm:"column:status;type:tinyint(4);not null"`
	ParamToDB string                  `gorm:"column:param"`
	Param     map[string]string       `gorm:"-"`
}

func (t *Trigger) String() string {
	return fmt.Sprintf("Trigger{ID: %d, Name: %s, JobID: %d, ScheduleType: %d, ScheduleConf: %s,"+
		" MisfireStrategy: %s, FailRetryCount: %d, ExecuteTimeout: %s, TriggerLastTime: %s, TriggerNextTime: %s, Status: %s}",
		t.ID,
		t.Name,
		t.JobID,
		t.ScheduleType,
		t.ScheduleConf,
		t.MisfireStrategy,
		t.FailRetryCount,
		t.ExecuteTimeout,
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

// PrepareFire Trigger准备被fire时调用，目前只是更新下次fire时间，并检查本次fire时间是否正确
func (t *Trigger) PrepareFire() error {
	var err error
	t.TriggerLastTime = t.TriggerNextTime
	t.TriggerNextTime, err = t.NextExecutionTime()
	return err
}

func TriggersToString(triggers []*Trigger) string {
	var builder strings.Builder
	for i, trigger := range triggers {
		builder.WriteString(fmt.Sprintf("Trigger %d: %s\n", i+1, trigger.String()))
	}

	return builder.String()
}

func (t *Trigger) BeforeCreate(tx *gorm.DB) error {
	return t.prepareParam()
}

func (t *Trigger) BeforeUpdate(tx *gorm.DB) error {
	return t.prepareParam()
}

func (t *Trigger) AfterFind(tx *gorm.DB) error {
	return t.parseParam()
}

func (t *Trigger) prepareParam() error {
	if t.Param != nil && len(t.Param) != 0 {
		jsonData, err := json.Marshal(t.Param)
		if err != nil {
			return err
		}
		t.ParamToDB = string(jsonData)
	}
	return nil
}

func (t *Trigger) parseParam() error {
	if t.ParamToDB != "" {
		var paramMap map[string]string
		err := json.Unmarshal([]byte(t.ParamToDB), &paramMap)
		if err != nil {
			return err
		}
		t.Param = paramMap
	}
	return nil
}
