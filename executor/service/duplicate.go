package service

import "supernova/pkg/api"

type DuplicateService struct {
}

func NewDuplicateService() *DuplicateService {
	return &DuplicateService{}
}

func (s *DuplicateService) OnReceiveRunJobRequest(request *api.RunJobRequest) {

}

func (s *DuplicateService) OnStartExecute(task *Task) {

}

func (s *DuplicateService) OnDuplicateOnFireLogID(request *api.RunJobRequest) {

}

func (s *DuplicateService) OnFinishExecute(response *api.RunJobResponse) {

}

func (s *DuplicateService) CheckDuplicateExecute(onFireID uint) *api.RunJobResponse {
	return nil
}

var _ ExecuteListener = (*DuplicateService)(nil)
