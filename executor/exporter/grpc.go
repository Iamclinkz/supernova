package exporter

import (
	"net"
	"strconv"
	"supernova/executor/handler"
	"supernova/executor/service"
	"supernova/pkg/api/executor"
	"supernova/pkg/discovery"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
)

type GrpcExporter struct {
	grpcHandler *handler.GrpcHandler
	grpcServer  server.Server
	serviceConf *discovery.ServiceServeConf
	stopCh      chan error
}

func NewGrpcExporter(executeService *service.ExecuteService,
	statisticsService *service.StatisticsService,
	serviceConf *discovery.ServiceServeConf,
	enableOTel bool) *GrpcExporter {
	e := new(GrpcExporter)
	e.grpcHandler = handler.NewGrpcHandler(executeService, statisticsService, enableOTel)
	e.stopCh = make(chan error)
	e.serviceConf = serviceConf
	return e
}

func (e *GrpcExporter) grpcExitChan() <-chan error {
	return e.stopCh
}

func (e *GrpcExporter) StartServe() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(e.serviceConf.Port))
	if err != nil {
		panic(err)
	}
	e.grpcServer = executor.NewServer(e.grpcHandler,
		server.WithServiceAddr(addr),
		//server.WithSuite(tracing.NewServerSuite()),
		//server.WithMiddleware(middleware.PrintKitexRequestResponse),
		// GrpcExitChan 麻了。。调Executor优雅退出的时候，发现grpc先断了。。调了好久才发现这里还要传个这个。。
		server.WithExitSignal(e.grpcExitChan),
	)
	klog.Infof("executor try start serve, protoc:grpc, port:%v", e.serviceConf.Port)
	if err := e.grpcServer.Run(); err != nil {
		//panic(err)
		//更新优雅退出后，如果这里返回错误，可能是优雅推出关闭grpc了。不需要panic
		klog.Errorf("grpc server stopped, error:%v", err)
	}
}

func (e *GrpcExporter) Stop() {
	_ = e.grpcServer.Stop()
	klog.Infof("GrpcExporter stopped")
}

func (e *GrpcExporter) GracefulStop() {
	e.grpcHandler.OnGracefulStop()
}

var _ Exporter = (*GrpcExporter)(nil)
