package service

import "supernova/pkg/api"

type StatisticsService struct {
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
	return &api.HealthStatus{Workload: 20}
}

var _ ExecuteListener = (*StatisticsService)(nil)
