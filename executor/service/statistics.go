package service

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
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
}

func NewStatisticsService(enableOTel bool) *StatisticsService {
	ret := &StatisticsService{
		enableOTel: enableOTel,
	}
	if enableOTel {
		ret.tracer = otel.Tracer("StatisticsTracer")
	}

	return ret
}

func (s *StatisticsService) OnSendRunJobResponseSuccess(response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(-1)
}

func (s *StatisticsService) OnNoNeedSendResponse(request *api.RunJobRequest) {
	s.unFinishedRequestCount.Add(-1)
}

func (s *StatisticsService) OnReceiveRunJobRequest(request *api.RunJobRequest) {
	klog.Tracef("receive execute jobRequest, OnFireID:%v", request.OnFireLogID)
	s.unFinishedRequestCount.Add(1)
}

func (s *StatisticsService) OnOverTimeTaskExecuteSuccess(request *api.RunJobRequest, response *api.RunJobResponse) {
	s.unFinishedRequestCount.Add(1)
}

func (s *StatisticsService) OnStartExecute(jobRequest *api.RunJobRequest) {
	s.currentJobCount.Add(1)
}

func (s *StatisticsService) OnFinishExecute(jobRequest *api.RunJobRequest, jobResponse *api.RunJobResponse) {
	s.currentJobCount.Add(-1)
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
