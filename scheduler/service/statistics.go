package service

import (
	"supernova/pkg/util"
	"time"
)

type StatisticsService struct {
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) GetHandleTriggerDuration() time.Duration {
	return time.Second * 15
}

func (s *StatisticsService) GetHandleTriggerForwardDuration() time.Duration {
	//todo 测试使用
	return time.Since(util.VeryEarlyTime())
}

func (s *StatisticsService) GetHandleTriggerMaxCount() int {
	return 3000
}

func (s *StatisticsService) GetHandleTimeoutOnFireLogMaxCount() int {
	return 3000
}

func (s *StatisticsService) GetScheduleInterval() time.Duration {
	return time.Second * 4
}

func (s *StatisticsService) GetCheckTimeoutOnFireLogsInterval() time.Duration {
	return time.Second * 5
}

func (s *StatisticsService) GetExecutorHeartbeatInterval() time.Duration {
	return time.Second * 10
}

func (s *StatisticsService) GetHeartBeatTimeout() time.Duration {
	return time.Second * 2
}
