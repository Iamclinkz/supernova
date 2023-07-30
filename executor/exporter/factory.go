package exporter

import (
	"supernova/executor/service"
	"supernova/pkg/discovery"
)

type Exporter interface {
	StartServe()
	Stop()
	GracefulStop()
}

func NewExporter(executeService *service.ExecuteService,
	statisticsService *service.StatisticsService,
	serviceConf *discovery.ExecutorServiceServeConf,
	enableOTel bool) Exporter {
	switch serviceConf.Protoc {
	case discovery.ProtocTypeGrpc:
		return NewGrpcExporter(executeService, statisticsService, serviceConf, enableOTel)
	default:
		//todo
		panic("")
	}
}
