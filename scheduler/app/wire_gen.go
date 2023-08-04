// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
	"go.opentelemetry.io/otel/sdk/metric"
	trace2 "go.opentelemetry.io/otel/sdk/trace"
	"supernova/pkg/discovery"
	"supernova/pkg/session/trace"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

// Injectors from wire.go:

func genScheduler(instanceID string, OTelConfig *trace.OTelConfig, tracerProvider *trace2.TracerProvider, meterProvider *metric.MeterProvider, scheduleOperator schedule_operator.Operator, client discovery.DiscoverClient, schedulerWorkerCount int, standalone bool) (*Scheduler, error) {
	statisticsService := service.NewStatisticsService(instanceID, OTelConfig, client, standalone)
	jobService := service.NewJobService(scheduleOperator, statisticsService)
	triggerService := service.NewTriggerService(scheduleOperator, statisticsService, jobService, standalone, instanceID)
	onFireService := service.NewOnFireService(scheduleOperator, statisticsService, instanceID)
	executorManageService := service.NewExecutorManageService(statisticsService, client)
	executorRouteService := service.NewExecutorRouteService(executorManageService)
	scheduleService := service.NewScheduleService(statisticsService, jobService, triggerService, onFireService, executorRouteService, schedulerWorkerCount, executorManageService, OTelConfig, standalone)
	scheduler := newSchedulerInner(instanceID, standalone, OTelConfig, scheduleOperator, client, tracerProvider, meterProvider, scheduleService, statisticsService, executorRouteService, executorManageService, jobService, triggerService, onFireService)
	return scheduler, nil
}
