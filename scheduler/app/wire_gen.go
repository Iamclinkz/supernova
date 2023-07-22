// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"supernova/pkg/discovery"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

// Injectors from wire.go:

func genScheduler(scheduleOperator schedule_operator.Operator, client discovery.Client, schedulerWorkerCount int) (*Scheduler, error) {
	statisticsService := service.NewStatisticsService()
	jobService := service.NewJobService(scheduleOperator, statisticsService)
	triggerService := service.NewTriggerService(scheduleOperator, statisticsService)
	onFireService := service.NewOnFireService(scheduleOperator, statisticsService)
	executorManageService := service.NewExecutorManageService(statisticsService, client)
	executorRouteService := service.NewExecutorRouteService(executorManageService)
	scheduleService := service.NewScheduleService(statisticsService, jobService, triggerService, onFireService, executorRouteService, schedulerWorkerCount, executorManageService)
	scheduler := newSchedulerInner(scheduleOperator, scheduleService, statisticsService, executorRouteService, executorManageService, jobService, triggerService, onFireService)
	return scheduler, nil
}
