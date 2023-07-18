package service

import (
	"supernova/scheduler/util"
	"time"
)

type StatisticsService struct {
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) GetHandleTriggerDuration() time.Duration {
	return time.Second * 3
}

func (s *StatisticsService) GetHandleTriggerForwardDuration() time.Duration {
	//todo 测试使用
	return time.Since(util.VeryEarlyTime())
}

func (s *StatisticsService) GetHandleTriggerMaxCount() int {
	return 100
}

func (s *StatisticsService) GetScheduleInterval() time.Duration {
	return time.Second * 2
}

func (s *StatisticsService) GetExecutorHeartbeatInterval() time.Duration {
	return time.Second * 10
}

func (s *StatisticsService) GetHeartBeatTimeout() time.Duration {
	return time.Second * 2
}
