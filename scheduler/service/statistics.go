package service

import (
	"context"
	"math/rand"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	trace2 "supernova/pkg/session/trace"
	"supernova/pkg/util"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type StatisticsService struct {
	instanceID string
	oTelConfig *trace2.OTelConfig
	tracer     trace.Tracer
	standalone bool

	discoveryClient       discovery.DiscoverClient
	currentSchedulerCount int
	shutdownCh            chan struct{}

	//动态调整本次拿取过期OnFireLog
	lastTimeFetchOnFireLogFailRate float32
	lastTimeFetchOnFireLogInterval time.Duration

	meter                         metric.Meter
	defaultMetricsOption          metric.MeasurementOption
	fetchedTriggersCounter        metric.Int64Counter   //获取待执行Trigger的个数
	foundTimeoutLogsCounter       metric.Int64Counter   //查找到超时的OnFireLog记录的个数
	holdTimeoutLogsFailureCounter metric.Int64Counter   //获取失败（乐观锁冲突）的OnFireLog记录的个数
	holdTimeoutLogsSuccessCounter metric.Int64Counter   //获取成功（乐观锁成功）的OnFireLog记录的个数
	fireFailureCounter            metric.Int64Counter   //fire失败的任务的个数
	fireSuccessCounter            metric.Int64Counter   //fire成功的任务的个数
	scheduleDelayHistogram        metric.Int64Histogram //调度时间（任务应该被安排执行时间 - 任务实际被安排执行时间）
}

func NewStatisticsService(instanceID string, oTelConfig *trace2.OTelConfig,
	discoveryClient discovery.DiscoverClient, standalone bool) *StatisticsService {
	ret := &StatisticsService{
		instanceID:                     instanceID,
		oTelConfig:                     oTelConfig,
		currentSchedulerCount:          1,
		discoveryClient:                discoveryClient,
		shutdownCh:                     make(chan struct{}),
		lastTimeFetchOnFireLogInterval: 4 * time.Second,
		standalone:                     standalone,
	}
	if oTelConfig.EnableTrace {
		ret.tracer = otel.Tracer("StatisticTracer")
	}

	if oTelConfig.EnableMetrics {
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

		ret.scheduleDelayHistogram, err = ret.meter.Int64Histogram("schedule_delay",
			metric.WithDescription("The difference between the scheduled execution time and the actual allocation time"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func (s *StatisticsService) WatchScheduler() {
	if s.standalone {
		klog.Infof("WatchScheduler stopped in standalone mode")
		return
	}

	for {
		select {
		case <-s.shutdownCh:
			return
		default:
			tmp := len(s.discoveryClient.DiscoverServices(constance.SchedulerServiceName))
			if tmp == 0 {
				tmp = 1
			}
			s.currentSchedulerCount = len(s.discoveryClient.DiscoverServices(constance.SchedulerServiceName))
			klog.Tracef("current scheduler count:%v", s.currentSchedulerCount)
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *StatisticsService) Stop() {
	close(s.shutdownCh)
}

// OnFetchNeedFireTriggers 获得的需要执行的Trigger的个数
func (s *StatisticsService) OnFetchNeedFireTriggers(count int) {
	if s.oTelConfig.EnableMetrics {
		s.fetchedTriggersCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnFindTimeoutOnFireLogs 找到的过期的OnFireLogs个数
func (s *StatisticsService) OnFindTimeoutOnFireLogs(count int) {
	if s.oTelConfig.EnableMetrics {
		s.foundTimeoutLogsCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnHoldTimeoutOnFireLogFail 获得失败的过期的OnFireLogs个数
func (s *StatisticsService) OnHoldTimeoutOnFireLogFail(count int) {
	if s.oTelConfig.EnableMetrics {
		s.holdTimeoutLogsFailureCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// OnHoldTimeoutOnFireLogSuccess 获得的过期的OnFireLogs个数
func (s *StatisticsService) OnHoldTimeoutOnFireLogSuccess(count int) {
	if s.oTelConfig.EnableMetrics {
		s.holdTimeoutLogsSuccessCounter.Add(context.Background(), int64(count), s.defaultMetricsOption)
	}
}

// UpdateLastTimeFetchTimeoutOnFireLogFailRate 更新上次失败率
func (s *StatisticsService) UpdateLastTimeFetchTimeoutOnFireLogFailRate(rate float32) {
	if rate > 1 || rate < 0 {
		panic("")
	}

	s.lastTimeFetchOnFireLogFailRate = rate
}

type FireFailReason string

const (
	FireFailReasonNoExecutor           FireFailReason = "NoExecutor"
	FireFailReasonExecutorConnectError FireFailReason = "ExecutorConnectFail"
	FireFailReasonUpdateOnFireLogError FireFailReason = "UpdateOnFireLogFail"
	FireFailReasonFindJobError         FireFailReason = "CanNotFindJob"
)

func (s *StatisticsService) GetHandleTriggerDuration() time.Duration {
	return time.Second * 15
}

func (s *StatisticsService) GetHandleTriggerForwardDuration() time.Duration {
	//todo 测试使用
	return time.Since(util.VeryEarlyTime())
}

func (s *StatisticsService) GetHandleTimeoutOnFireLogMaxCount() int {
	if s.standalone {
		return 10000
	}
	return 3000
}

func (s *StatisticsService) GetScheduleInterval() time.Duration {
	return time.Second * 1
}

func (s *StatisticsService) GetCheckTimeoutOnFireLogsInterval() time.Duration {
	var (
		//CheckTimeoutOnFireLogsMaxInterval 每多一个Scheduler实例，多加0.5的最大上限
		CheckTimeoutOnFireLogsMaxInterval = 200*time.Millisecond + 500*time.Millisecond*time.Duration(s.currentSchedulerCount)
		CheckTimeoutOnFireLogsMinInterval = 100 * time.Millisecond
	)

	// 根据冲突率调整间隔
	var adjustedInterval time.Duration
	if s.lastTimeFetchOnFireLogFailRate <= 0.2 || s.currentSchedulerCount <= 1 {
		//如果冲突率在0.2以下，或者没有别的scheduler，适当缩短间隔
		adjustedInterval = time.Duration(float64(s.lastTimeFetchOnFireLogInterval) * 0.8)
	} else if s.lastTimeFetchOnFireLogFailRate > 0.2 && s.lastTimeFetchOnFireLogFailRate < 0.5 {
		//如果冲突率在0.4到0.7之间，增加50%的间隔
		adjustedInterval = time.Duration(float64(s.lastTimeFetchOnFireLogInterval) * 1.5)
	} else {
		//如果冲突率在0.5以上，增加100%的间隔
		adjustedInterval = s.lastTimeFetchOnFireLogInterval * 2
	}

	//确保间隔在最小和最大值之间
	if adjustedInterval < CheckTimeoutOnFireLogsMinInterval {
		adjustedInterval = CheckTimeoutOnFireLogsMinInterval
	} else if adjustedInterval > CheckTimeoutOnFireLogsMaxInterval {
		adjustedInterval = CheckTimeoutOnFireLogsMaxInterval
	}

	//在计算出的间隔基础上加入一个随机抖动，范围为[-25%, +25%]，以降低冲突概率
	jitterFactor := 0.25
	jitter := time.Duration(rand.Float64()*float64(adjustedInterval)*jitterFactor*2 - float64(adjustedInterval)*jitterFactor)
	adjustedInterval += jitter

	s.lastTimeFetchOnFireLogInterval = adjustedInterval
	klog.Errorf("[%v]  %v  -> %v", s.instanceID, s.lastTimeFetchOnFireLogFailRate, adjustedInterval)
	return adjustedInterval
}

func (s *StatisticsService) GetExecutorHeartbeatInterval() time.Duration {
	return time.Second * 5
}

func (s *StatisticsService) GetCheckExecutorHeartBeatTimeout() time.Duration {
	return time.Second * 2
}

// RecordScheduleDelay 记录任务调度的时间。即任务应该被安排执行时间 - 任务实际被安排执行时间
func (s *StatisticsService) RecordScheduleDelay(delay time.Duration) {
	if !s.oTelConfig.EnableMetrics {
		return
	}

	//klog.Errorf("[%v]:%v", s.instanceID, delay)
	s.scheduleDelayHistogram.Record(context.Background(), delay.Milliseconds(), s.defaultMetricsOption)
}

// GetHandleTriggerMaxCount 获取本次最多获取多少条待触发的Trigger
func (s *StatisticsService) GetHandleTriggerMaxCount() int {
	return 200000
}

// OnFireFail 任务扔给Executor执行失败
func (s *StatisticsService) OnFireFail(reason FireFailReason) {
	if s.oTelConfig.EnableMetrics {
		s.fireFailureCounter.Add(context.Background(), 1,
			metric.WithAttributes(attribute.String("reason", string(reason))), s.defaultMetricsOption)
	}
}

// OnFireSuccess 任务扔给Executor执行成功
func (s *StatisticsService) OnFireSuccess() {
	if s.oTelConfig.EnableMetrics {
		s.fireSuccessCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}
