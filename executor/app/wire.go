//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	sdkmetrics "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"supernova/executor/processor"
	"supernova/executor/service"
	"supernova/pkg/discovery"
	trace2 "supernova/pkg/session/trace"
)

func genExecutor(
	instanceID string,
	oTelConfig *trace2.OTelConfig,
	traceProvider *sdktrace.TracerProvider,
	meterProvider *sdkmetrics.MeterProvider,
	tags []string,
	processor map[string]processor.JobProcessor,
	serveConf *discovery.ServiceServeConf,
	processorCount int, client discovery.DiscoverClient,
	extraConf map[string]string,
) (*Executor, error) {
	wire.Build(
		newExecutorInner,

		//service
		service.NewDuplicateService,
		service.NewExecuteService,
		service.NewProcessorService,
		service.NewStatisticsService,
	)

	return &Executor{}, nil
}
