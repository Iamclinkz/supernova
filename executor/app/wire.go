//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"supernova/executor/processor"
	"supernova/executor/service"
	"supernova/pkg/discovery"
)

func genExecutor(
	instanceID string,
	tags []string,
	processor map[string]processor.JobProcessor,
	serveConf *discovery.ExecutorServiceServeConf,
	processorCount int, client discovery.ExecutorDiscoveryClient,
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
