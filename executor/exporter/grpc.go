package exporter

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	"net"
	"strconv"
	"supernova/executor/handler"
	"supernova/executor/service"
	"supernova/pkg/api/executor"
	"supernova/pkg/discovery"
)

type GrpcExporter struct {
	grpcHandler *handler.GrpcHandler
	grpcServer  server.Server
	serviceConf *discovery.ServiceServeConf
}

func NewGrpcExporter(executeService *service.ExecuteService,
	statisticsService *service.StatisticsService, serviceConf *discovery.ServiceServeConf) *GrpcExporter {
	e := new(GrpcExporter)
	e.grpcHandler = handler.NewGrpcHandler(executeService, statisticsService)
	e.serviceConf = serviceConf
	return e
}

func (e *GrpcExporter) StartServe() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(e.serviceConf.Port))
	if err != nil {
		panic(err)
	}
	e.grpcServer = executor.NewServer(e.grpcHandler,
		server.WithServiceAddr(addr),
		server.WithSuite(tracing.NewServerSuite()),
		//server.WithMiddleware(middleware.PrintKitexRequestResponse),
	)
	klog.Infof("executor try start serve, protoc:grpc, port:%v", e.serviceConf.Port)
	if err := e.grpcServer.Run(); err != nil {
		panic(err)
	}
}

func (e *GrpcExporter) Stop() {
	_ = e.grpcServer.Stop()
}

func (e *GrpcExporter) GracefulStop() {
	e.grpcHandler.OnGracefulStop()
}

var _ Exporter = (*GrpcExporter)(nil)