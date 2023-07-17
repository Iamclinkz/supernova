package model

import (
	"supernova/pkg/api"
)

type ExecutorStatus struct {
	RunningJob int
	Cpu        int
	Memory     int
}

func (s *ExecutorStatus) FromPb(response *api.HeartBeatResponse) {
	s.RunningJob = int(response.Cpu)
	s.Memory = int(response.Memory)
	s.RunningJob = int(response.RunningJob)
}
