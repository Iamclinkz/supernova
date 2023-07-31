package service

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"supernova/pkg/api"
	"sync/atomic"
)

type StatisticsService struct {
	gracefulStopped atomic.Bool
	currentJobCount atomic.Int64 //当前正在执行的job数
	//当前还没有处理结束的response个数。这里的“还没有处理结束”指的是如果需要回复，那么回复成功（stream发送成功）算结束
	//如果不需要回复，那么直接结束。
	unFinishedRequestCount atomic.Int64
	enableOTel             bool
	tracer                 trace.Tracer

	meter                               metric.Meter
	currentJobCountUpDownCounter        metric.Int64UpDownCounter
	unFinishedRequestCountUpDownCounter metric.Int64UpDownCounter
	runJobResponseSuccessCounter        metric.Int64Counter
	noNeedSendResponseCounter           metric.Int64Counter
	receiveRunJobRequestCounter         metric.Int64Counter
	overTimeTaskSuccessCounter          metric.Int64Counter
}

func NewStatisticsService(enableOTel bool) *StatisticsService {
	ret := &StatisticsService{
		enableOTel: enableOTel,
	}
	if enableOTel {
		ret.tracer = otel.Tracer("StatisticsTracer")
		ret.meter = otel.Meter("StatisticsMeter")

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

		ret.noNeedSendResponseCounter, err = ret.meter.Int64Counter("no_need_send_response_total",
			metric.WithDescription("Total number of requests with no need to send a response"))
		if err != nil {
			panic(err)
		}

		ret.receiveRunJobRequestCounter, err = ret.meter.Int64Counter("receive_run_job_request_total",
			metric.WithDescription("Total number of received run job requests"))
		if err != nil {
			panic(err)
		}

		ret.overTimeTaskSuccessCounter, err = ret.meter.Int64Counter("over_time_task_success_total",
			metric.WithDescription("Total number of successful over time task executions"))
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func (s *StatisticsService) OnSendRunJobResponseSuccess(response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(-1)
	if s.enableOTel {
		s.runJobResponseSuccessCounter.Add(context.Background(), 1)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), -1)
	}
}

func (s *StatisticsService) OnNoNeedSendResponse(request *api.RunJobRequest) {
	s.unFinishedRequestCount.Add(-1)
	if s.enableOTel {
		s.noNeedSendResponseCounter.Add(context.Background(), 1)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), -1)
	}
}

func (s *StatisticsService) OnReceiveRunJobRequest(request *api.RunJobRequest) {
	klog.Tracef("receive execute jobRequest, OnFireID:%v", request.OnFireLogID)
	s.unFinishedRequestCount.Add(1)
	if s.enableOTel {
		s.receiveRunJobRequestCounter.Add(context.Background(), 1)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), 1)
	}
}

func (s *StatisticsService) OnOverTimeTaskExecuteSuccess(request *api.RunJobRequest, response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(1)
	if s.enableOTel {
		s.overTimeTaskSuccessCounter.Add(context.Background(), 1)
		s.unFinishedRequestCountUpDownCounter.Add(context.Background(), 1)
	}
}

func (s *StatisticsService) OnStartExecute(jobRequest *api.RunJobRequest) {
	s.currentJobCount.Add(1)
	if s.enableOTel {
		s.currentJobCountUpDownCounter.Add(context.Background(), 1)
	}
}

func (s *StatisticsService) OnFinishExecute(jobRequest *api.RunJobRequest, jobResponse *api.RunJobResponse) {
	s.currentJobCount.Add(-1)
	if s.enableOTel {
		s.currentJobCountUpDownCounter.Add(context.Background(), -1)
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
