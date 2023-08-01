package service

import (
	"context"
	"math/rand"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
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
	enableOTel bool
	tracer     trace.Tracer

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

func NewStatisticsService(instanceID string, enableOTel bool, discoveryClient discovery.DiscoverClient) *StatisticsService {
	ret := &StatisticsService{
		instanceID:                     instanceID,
		enableOTel:                     enableOTel,
		currentSchedulerCount:          1,
		discoveryClient:                discoveryClient,
		shutdownCh:                     make(chan struct{}),
		lastTimeFetchOnFireLogInterval: 4 * time.Second,
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
	watchSchedulerTicker := time.NewTicker(1 * time.Millisecond)

	for {
		select {
		case <-s.shutdownCh:
			watchSchedulerTicker.Stop()
			return
		case <-watchSchedulerTicker.C:
			s.currentSchedulerCount = len(s.discoveryClient.DiscoverServices(constance.ExecutorServiceName))
			klog.Errorf("current scheduler count:%v", s.currentSchedulerCount)
			watchSchedulerTicker.Reset(time.Second * 2)
		}
	}
}

func (s *StatisticsService) Stop() {
	close(s.shutdownCh)
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
	return 5000
}

func (s *StatisticsService) GetScheduleInterval() time.Duration {
	return time.Second * 2
}

func (s *StatisticsService) GetCheckTimeoutOnFireLogsInterval() time.Duration {
	var (
		//CheckTimeoutOnFireLogsMaxInterval 每多一个Scheduler实例，多加2s的最大上限
		CheckTimeoutOnFireLogsMaxInterval = 4*time.Second + 2*time.Duration(s.currentSchedulerCount)
		CheckTimeoutOnFireLogsMinInterval = 2 * time.Second
	)

	// 根据冲突率调整间隔
	var adjustedInterval time.Duration
	if s.lastTimeFetchOnFireLogFailRate <= 0.4 || s.currentSchedulerCount <= 1 {
		//如果冲突率在0.4以下，或者没有别的scheduler，适当缩短间隔
		adjustedInterval = time.Duration(float64(s.lastTimeFetchOnFireLogInterval) * 0.8)
	} else if s.lastTimeFetchOnFireLogFailRate > 0.4 && s.lastTimeFetchOnFireLogFailRate < 0.7 {
		//如果冲突率在0.4到0.7之间，增加50%的间隔
		adjustedInterval = time.Duration(float64(s.lastTimeFetchOnFireLogInterval) * 1.5)
	} else {
		//如果冲突率在0.7以上，增加100%的间隔
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
	if !s.enableOTel {
		return
	}

	//klog.Errorf("[%v]:%v", s.instanceID, delay)
	s.scheduleDelayHistogram.Record(context.Background(), delay.Milliseconds(), s.defaultMetricsOption)
}

// GetHandleTriggerMaxCount 获取本次最多获取多少条待触发的Trigger
func (s *StatisticsService) GetHandleTriggerMaxCount() int {
	return 15000
}

// OnFireFail 任务扔给Executor执行失败
func (s *StatisticsService) OnFireFail(reason FireFailReason) {
	if s.enableOTel {
		s.fireFailureCounter.Add(context.Background(), 1,
			metric.WithAttributes(attribute.String("reason", string(reason))), s.defaultMetricsOption)
	}
}

// OnFireSuccess 任务扔给Executor执行成功
func (s *StatisticsService) OnFireSuccess() {
	if s.enableOTel {
		s.fireSuccessCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}
