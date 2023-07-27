package model

import (
	"supernova/pkg/api"
)

type ExecutorStatus struct {
	Workload        float32
	GracefulStopped bool
}

func NewExecutorStatusFromPb(response *api.HeartBeatResponse) *ExecutorStatus {
	return &ExecutorStatus{
		Workload:        response.HealthStatus.Workload,
		GracefulStopped: response.HealthStatus.GracefulStopped,
	}
}
