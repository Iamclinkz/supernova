//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"supernova/pkg/discovery"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

func genScheduler(scheduleOperator schedule_operator.Operator, client discovery.Client) (*Scheduler, error) {
	wire.Build(
		newSchedulerInner,

		//service
		service.NewExecutorManageService,
		service.NewExecutorRouteService,
		service.NewJobService,
		service.NewOnFireService,
		service.NewScheduleService,
		service.NewStatisticsService,
		service.NewTriggerService,
	)

	return &Scheduler{}, nil
}
