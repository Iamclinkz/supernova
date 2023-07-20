package service

import "supernova/pkg/api"

type StatisticsService struct {
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) OnReceiveRunJobRequest(request *api.RunJobRequest) {

}

func (s *StatisticsService) OnStartExecute(task *Task) {

}

func (s *StatisticsService) OnDuplicateOnFireLogID(request *api.RunJobRequest) {

}

func (s *StatisticsService) OnFinishExecute(response *api.RunJobResponse) {

}

func (s *StatisticsService) GetStatus() *api.HealthStatus {
	return &api.HealthStatus{Workload: 20}
}

var _ ExecuteListener = (*StatisticsService)(nil)
