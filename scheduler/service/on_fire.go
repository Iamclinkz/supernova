package service

import (
	"context"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
)

type OnFireService struct {
	jobOperator       schedule_operator.Operator
	statisticsService *StatisticsService
}

func NewOnFireService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService) *OnFireService {
	return &OnFireService{
		jobOperator:       jobOperator,
		statisticsService: statisticsService,
	}
}

func (s *OnFireService) DeleteOnFireLogFromID(ctx context.Context, onFireLogID uint) error {
	return s.jobOperator.DeleteOnFireLogFromID(ctx, onFireLogID)
}

func (s *JobService) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	return s.jobOperator.UpdateOnFireLogExecutorStatus(ctx, onFireLog)
}
