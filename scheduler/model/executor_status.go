package model

import (
	"supernova/pkg/api"
)

type ExecutorStatus struct {
	RunningJob int
	Cpu        int
	Memory     int
	Workload   float32
}

func (s *ExecutorStatus) FromPb(response *api.HeartBeatResponse) {
	s.Workload = response.HealthStatus.Workload
}
