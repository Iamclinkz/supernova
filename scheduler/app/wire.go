//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	sdkmetrics "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"supernova/pkg/discovery"
	"supernova/pkg/session/trace"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

func genScheduler(instanceID string, OTelConfig *trace.OTelConfig,
	tracerProvider *sdktrace.TracerProvider, meterProvider *sdkmetrics.MeterProvider,
	scheduleOperator schedule_operator.Operator, client discovery.DiscoverClient, schedulerWorkerCount int, standalone bool) (*Scheduler, error) {
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
