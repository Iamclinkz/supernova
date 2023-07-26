package service

import (
	"supernova/pkg/api"
	"sync/atomic"
)

type StatisticsService struct {
	gracefulStopped     atomic.Bool
	currentJobCount     atomic.Int64 //当前正在执行的job数
	unReplyRequestCount atomic.Int64 //当前还没有回复的response个数
}

func (s *StatisticsService) OnSendRunJobResponse(response *api.RunJobResponse) {

}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) OnReceiveRunJobRequest(request *api.RunJobRequest) {

}

func (s *StatisticsService) OnStartExecute(jobRequest *api.RunJobRequest) {

}

func (s *StatisticsService) OnFinishExecute(jobRequest *api.RunJobRequest, jobResponse *api.RunJobResponse) {

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
	return s.unReplyRequestCount.Load()
}
