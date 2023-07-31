package service

import (
	"context"
	"supernova/pkg/util"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type StatisticsService struct {
	instanceID string
	enableOTel bool
	tracer     trace.Tracer

	meter                         metric.Meter
	defaultMetricsOption          metric.MeasurementOption
	fetchedTriggersCounter        metric.Int64Counter
	foundTimeoutLogsCounter       metric.Int64Counter
	holdTimeoutLogsFailureCounter metric.Int64Counter
	holdTimeoutLogsSuccessCounter metric.Int64Counter
	fireFailureCounter            metric.Int64Counter
	fireSuccessCounter            metric.Int64Counter
	scheduleDelayHistogram        metric.Int64Histogram
}

func NewStatisticsService(instanceID string, enableOTel bool) *StatisticsService {
	ret := &StatisticsService{
		instanceID: instanceID,
		enableOTel: enableOTel,
	}
	if enableOTel {
		ret.tracer = otel.Tracer("StatisticTracer")
		ret.meter = otel.Meter("StatisticMeter")
		ret.defaultMetricsOption = metric.WithAttributes(
			attribute.Key("InstanceID").String(instanceID),
		)

		var err error
		ret.fetchedTriggersCounter, err = ret.meter.Int64Counter("fetched_triggers_total",
			metric.WithDescription("Total number of fetched triggers"))
		if err != nil {
			panic(err)
		}

		ret.foundTimeoutLogsCounter, err = ret.meter.Int64Counter("found_timeout_logs_total",
			metric.WithDescription("Total number of found timeout logs"))
		if err != nil {
			panic(err)
		}

		ret.holdTimeoutLogsFailureCounter, err = ret.meter.Int64Counter("hold_timeout_logs_failure_total",
			metric.WithDescription("Total number of hold timeout logs failures"))
		if err != nil {
			panic(err)
		}

		ret.holdTimeoutLogsSuccessCounter, err = ret.meter.Int64Counter("hold_timeout_logs_success_total",
			metric.WithDescription("Total number of hold timeout logs success"))
		if err != nil {
			panic(err)
		}

		ret.fireFailureCounter, err = ret.meter.Int64Counter("fire_failure_total",
			metric.WithDescription("Total number of fire failures"))
		if err != nil {
			panic(err)
		}

		ret.fireSuccessCounter, err = ret.meter.Int64Counter("fire_success_total",
			metric.WithDescription("Total number of fire successes"))
		if err != nil {
			panic(err)
		}

		// 创建Histogram指标，并设置桶的边界
		ret.scheduleDelayHistogram, err = ret.meter.Int64Histogram("scheduler.schedule_delay",
			metric.WithDescription("The difference between the scheduled execution time and the actual allocation time"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			panic(err)
		}
	}

	return ret
}

// OnFetchNeedFireTriggers 获得的需要执行的Trigger的个数
func (s *StatisticsService) OnFetchNeedFireTriggers(count int) {
	if s.enableOTel {
		s.fetchedTriggersCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnFindTimeoutOnFireLogs 找到的过期的OnFireLogs个数
func (s *StatisticsService) OnFindTimeoutOnFireLogs(count int) {
	if s.enableOTel {
		s.foundTimeoutLogsCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnHoldTimeoutOnFireLogFail 获得失败的过期的OnFireLogs个数
func (s *StatisticsService) OnHoldTimeoutOnFireLogFail(count int) {
	if s.enableOTel {
		s.holdTimeoutLogsFailureCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnHoldTimeoutOnFireLogSuccess 获得的过期的OnFireLogs个数
func (s *StatisticsService) OnHoldTimeoutOnFireLogSuccess(count int) {
	if s.enableOTel {
		s.holdTimeoutLogsSuccessCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

type FireFailReason string

const (
	FireFailReasonNoExecutor           FireFailReason = "NoExecutor"
	FireFailReasonExecutorConnectError FireFailReason = "ExecutorConnectFail"
	FireFailReasonUpdateOnFireLogError FireFailReason = "UpdateOnFireLogFail"
	FireFailReasonFindJobError         FireFailReason = "CanNotFindJob"
)

func (s *StatisticsService) OnFireFail(reason FireFailReason) {
	if s.enableOTel {
		s.fireFailureCounter.Add(context.Background(), 1,
			metric.WithAttributes(attribute.String("reason", string(reason))), s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnFireSuccess() {
	if s.enableOTel {
		s.fireSuccessCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) GetHandleTriggerDuration() time.Duration {
	return time.Second * 15
}

func (s *StatisticsService) GetHandleTriggerForwardDuration() time.Duration {
	//todo 测试使用
	return time.Since(util.VeryEarlyTime())
}

func (s *StatisticsService) GetHandleTriggerMaxCount() int {
	return 6000
}

func (s *StatisticsService) GetHandleTimeoutOnFireLogMaxCount() int {
	return 3000
}

func (s *StatisticsService) GetScheduleInterval() time.Duration {
	return time.Second * 3
}

func (s *StatisticsService) GetCheckTimeoutOnFireLogsInterval() time.Duration {
	return time.Second * 5
}

func (s *StatisticsService) GetExecutorHeartbeatInterval() time.Duration {
	return time.Second * 5
}

func (s *StatisticsService) GetCheckExecutorHeartBeatTimeout() time.Duration {
	return time.Second * 2
}

func (s *StatisticsService) RecordScheduleDelay(delay time.Duration) {
	if !s.enableOTel {
		return
	}

	// 将时间差值转换为毫秒
	delayMs := float64(delay) / float64(time.Millisecond)

	// 更新Histogram指标
	s.scheduleDelayHistogram.Record(context.Background(), int64(delayMs), s.defaultMetricsOption)
}
