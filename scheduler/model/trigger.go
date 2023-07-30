package model

import (
	"fmt"
	"github.com/google/btree"
	"strings"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"time"

	"github.com/robfig/cron/v3"
)

type Trigger struct {
	ID        uint
	UpdatedAt time.Time
	Name      string
	//关联的任务ID
	JobID uint
	// 选择调度的策略（None表示不触发，Cron表示使用Cron表达式，
	ScheduleType constance.ScheduleType
	// 配合调度策略，这里的配置信息
	ScheduleConf string
	// 如果错过了本次执行，应该采取的策略。和Fail-Retry概念不一样，misfire指任务本身没有发生错误，但是
	// 例如Scheduler有太多任务做不过来等原因，使得调度到本trigger的时候，已经错过了本trigger的执行时间。
	//todo 待实现
	MisfireStrategy constance.MisfireStrategyType
	// 如果出错，最大重试次数。
	FailRetryCount int
	// executor执行的最大时间。要求用户一定指定。如果超时，则减少OnFireLog中的RetryCount字段
	ExecuteTimeout    time.Duration
	FailRetryInterval time.Duration //失败重试间隔，为0则立刻重试
	// 上次触发时间
	TriggerLastTime time.Time
	// 下次触发时间
	TriggerNextTime time.Time
	// 当前状态
	Status      constance.TriggerStatus
	Param       map[string]string
	AtLeastOnce bool //语义，至少一次。如果为false，则为至多一次
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

func (t *Trigger) Less(than btree.Item) bool {
	return t.TriggerNextTime.Before(than.(*Trigger).TriggerNextTime)
}
