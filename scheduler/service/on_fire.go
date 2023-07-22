package service

import (
	"context"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
)

type OnFireService struct {
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService
}

func NewOnFireService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService) *OnFireService {
	return &OnFireService{
		scheduleOperator:  jobOperator,
		statisticsService: statisticsService,
	}
}
func (s *OnFireService) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	return s.scheduleOperator.UpdateOnFireLogExecutorStatus(ctx, onFireLog)
}

func (s *OnFireService) UpdateOnFireLogFail(ctx context.Context, onFireLog uint, errorMsg string) error {
	return s.scheduleOperator.UpdateOnFireLogFail(ctx, onFireLog, errorMsg)
}

func (s *OnFireService) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	return s.scheduleOperator.UpdateOnFireLogSuccess(ctx, onFireLogID, result)
}

func (s *OnFireService) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	return s.scheduleOperator.UpdateOnFireLogStop(ctx, onFireLogID, msg)
}
