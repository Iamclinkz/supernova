//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"supernova/pkg/discovery"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

func genScheduler(instanceID string, enableTrace bool, traceProvider provider.OtelProvider, scheduleOperator schedule_operator.Operator, client discovery.ExecutorDiscoveryClient, schedulerWorkerCount int) (*Scheduler, error) {
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
