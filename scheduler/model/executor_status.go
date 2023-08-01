package model

import (
	"fmt"
	"supernova/pkg/api"
)

type ExecutorStatus struct {
	Workload        float32
	GracefulStopped bool
}

func (s *ExecutorStatus) String() string {
	return fmt.Sprintf("Workload:%v, GracefulStopped:%v", s.Workload, s.GracefulStopped)
}

func NewExecutorStatusFromPb(response *api.HeartBeatResponse) *ExecutorStatus {
	return &ExecutorStatus{
		Workload:        response.HealthStatus.Workload,
		GracefulStopped: response.HealthStatus.GracefulStopped,
	}
}
