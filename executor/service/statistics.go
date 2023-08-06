package service

import (
	"context"
	"supernova/pkg/api"
	trace2 "supernova/pkg/session/trace"
	"supernova/pkg/util"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type StatisticsService struct {
	instanceID      string
	gracefulStopped atomic.Bool
	currentJobCount atomic.Int64 //当前正在执行的job数
	//当前还没有处理结束的response个数。这里的“还没有处理结束”指的是如果需要回复，那么回复成功（stream发送成功）算结束
	//如果不需要回复，那么直接结束。
	unFinishedRequestCount atomic.Int64
	oTelConfig             *trace2.OTelConfig

	//trace
	tracer trace.Tracer

	//metrics
	meter                               metric.Meter
	defaultMetricsOption                metric.MeasurementOption
	currentJobCountUpDownCounter        metric.Int64UpDownCounter //当前正在（并发）执行的任务的数量
	unFinishedRequestCountUpDownCounter metric.Int64UpDownCounter //当前未处理的请求的数量
	runJobResponseSuccessCounter        metric.Int64Counter       //发送执行任务结果成功数量
	duplicateExecuteRequestCounter      metric.Int64Counter       //重复执行（正在执行 or 已执行成功）请求数量
	receiveRunJobRequestCounter         metric.Int64Counter       //接收的RunJobRequest消息的数量
	overtimeTaskCounter                 metric.Int64Counter       //超时的任务的数量
	executeJobCount                     metric.Int64Counter       //执行任务的数量
	executeTimeHistogram                metric.Int64Histogram     //纯执行时间
	failJobCount                        metric.Int64Counter       //任务执行失败数
}

func NewStatisticsService(instanceID string, oTelConfig *trace2.OTelConfig) *StatisticsService {
	ret := &StatisticsService{
		oTelConfig: oTelConfig,
		instanceID: instanceID,
	}
	if oTelConfig.EnableTrace {
		//trace
		ret.tracer = otel.Tracer("StatisticsTracer")
	}

	if oTelConfig.EnableMetrics {
		//metrics
		ret.meter = otel.Meter("StatisticsMeter")
		ret.defaultMetricsOption = metric.WithAttributes(
			attribute.Key("InstanceID").String(instanceID),
			attribute.Key("env").String(util.GetEnv()),
		)

		var err error
		ret.currentJobCountUpDownCounter, err = ret.meter.Int64UpDownCounter("current_job_count",
			metric.WithDescription("Current number of executing jobs"))
		if err != nil {
			panic(err)
		}

		ret.unFinishedRequestCountUpDownCounter, err = ret.meter.Int64UpDownCounter("unfinished_request_count",
			metric.WithDescription("Current number of unfinished requests"))
		if err != nil {
			panic(err)
		}

		ret.runJobResponseSuccessCounter, err = ret.meter.Int64Counter("run_job_response_success_total",
			metric.WithDescription("Total number of successful run job responses sent"))
		if err != nil {
			panic(err)
		}

		ret.duplicateExecuteRequestCounter, err = ret.meter.Int64Counter("duplicate_execute_request_total",
			metric.WithDescription("Total number of requests with no need to send a response"))
		if err != nil {
			panic(err)
		}

		ret.receiveRunJobRequestCounter, err = ret.meter.Int64Counter("receive_run_job_request_total",
			metric.WithDescription("Total number of received run job requests"))
		if err != nil {
			panic(err)
		}

		ret.overtimeTaskCounter, err = ret.meter.Int64Counter("over_time_task_total",
			metric.WithDescription("Total number of task overtime"))
		if err != nil {
			panic(err)
		}

		ret.executeJobCount, err = ret.meter.Int64Counter("execute_job_total",
			metric.WithDescription("Total number of job execution"))
		if err != nil {
			panic(err)
		}

		ret.executeTimeHistogram, err = ret.meter.Int64Histogram("execute_time",
			metric.WithDescription("Execute time by executors"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			panic(err)
		}

		ret.failJobCount, err = ret.meter.Int64Counter("fail_job_count",
			metric.WithDescription("Execute fail job count"),
		)
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func (s *StatisticsService) OnSendRunJobResponseSuccess(response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(-1)
	if s.oTelConfig.EnableMetrics {
		s.runJobResponseSuccessCounter.Add(context.Background(), 1, s.defaultMetricsOption)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), -1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnNoNeedSendResponse(request *api.RunJobRequest) {
	s.unFinishedRequestCount.Add(-1)
	if s.oTelConfig.EnableMetrics {
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), -1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnDuplicateRequest(request *api.RunJobRequest) {
	if s.oTelConfig.EnableMetrics {
		s.duplicateExecuteRequestCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnReceiveRunJobRequest(request *api.RunJobRequest) {
	klog.Tracef("receive execute jobRequest, OnFireID:%v", request.OnFireLogID)
	s.unFinishedRequestCount.Add(1)
	if s.oTelConfig.EnableMetrics {
		s.receiveRunJobRequestCounter.Add(context.Background(), 1, s.defaultMetricsOption)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnOverTimeTaskExecuteSuccess(request *api.RunJobRequest, response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(1)
	if s.oTelConfig.EnableMetrics {
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnStartExecute(jobRequest *api.RunJobRequest) {
	s.currentJobCount.Add(1)
	if s.oTelConfig.EnableMetrics {
		s.currentJobCountUpDownCounter.Add(context.Background(), 1, s.defaultMetricsOption)
		s.executeJobCount.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) OnFinishExecute(jobRequest *api.RunJobRequest, jobResponse *api.RunJobResponse) {
	s.currentJobCount.Add(-1)
	if s.oTelConfig.EnableMetrics {
		s.currentJobCountUpDownCounter.Add(context.Background(), -1, s.defaultMetricsOption)
		if !jobResponse.Result.Ok {
			s.failJobCount.Add(context.TODO(), 1)
		}
	}
}

func (s *StatisticsService) OnTaskOvertime(jobRequest *api.RunJobRequest) {
	if s.oTelConfig.EnableMetrics {
		s.overtimeTaskCounter.Add(context.Background(), 1, s.defaultMetricsOption)
	}
}

func (s *StatisticsService) GetStatus() *api.HealthStatus {
	if s.gracefulStopped.Load() {
		return &api.HealthStatus{
			Workload:        100,
			GracefulStopped: true,
		}
	}
	return &api.HealthStatus{
		Workload:        20,
		GracefulStopped: false,
	}
}

var _ ExecuteListener = (*StatisticsService)(nil)

// OnGracefulStop 优雅退出，会将自己的健康上报改为0，从而让Scheduler不再给自己发任务
func (s *StatisticsService) OnGracefulStop() {
	s.gracefulStopped.Store(true)
}

func (s *StatisticsService) GetUnReplyRequestCount() int64 {
	return s.unFinishedRequestCount.Load()
}

// RecordExecuteTime 记录任务执行时间
func (s *StatisticsService) RecordExecuteTime(delay time.Duration) {
	if !s.oTelConfig.EnableMetrics {
		return
	}

	//	klog.Errorf("timeout:%v", delay)
	s.executeTimeHistogram.Record(context.Background(), delay.Milliseconds(), s.defaultMetricsOption)
}
