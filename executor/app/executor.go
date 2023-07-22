package app

import (
	"net"
	"strconv"
	myConstance "supernova/executor/constance"
	"supernova/executor/handler"
	"supernova/executor/processor"
	"supernova/executor/service"
	"supernova/pkg/api/executor"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/pkg/util"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

type Executor struct {
	//config
	instanceID     string
	tags           []string
	processor      map[string]processor.JobProcessor
	serveConf      *discovery.ServiceServeConf
	processorCount int
	extraConf      map[string]string

	//discovery
	discoveryClient discovery.Client

	//service
	duplicateService  *service.DuplicateService
	executeService    *service.ExecuteService
	processorService  *service.ProcessorService
	statisticsService *service.StatisticsService

	//grpc server
	grpcServer server.Server
}

func newExecutorInner(
	instanceID string,
	tags []string,
	processor map[string]processor.JobProcessor,
	serveConf *discovery.ServiceServeConf,
	processorCount int,
	extraConf map[string]string,

	discoveryClient discovery.Client,
	duplicateService *service.DuplicateService,
	executeService *service.ExecuteService,
	processorService *service.ProcessorService,
	statisticsService *service.StatisticsService,
) *Executor {
	ret := &Executor{
		instanceID:        instanceID,
		tags:              tags,
		processor:         processor,
		serveConf:         serveConf,
		extraConf:         extraConf,
		processorCount:    processorCount,
		discoveryClient:   discoveryClient,
		duplicateService:  duplicateService,
		executeService:    executeService,
		processorService:  processorService,
		statisticsService: statisticsService,
	}

	for _, p := range processor {
		ret.processorService.Register(p)
	}
	return ret
}

func (e *Executor) startServe() {
	switch e.serveConf.Protoc {
	case discovery.ProtocTypeGrpc:
		grpcHandler := handler.NewGrpcHandler(e.executeService, e.statisticsService)
		addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(e.serveConf.Port))
		if err != nil {
			panic(err)
		}

		svr := executor.NewServer(grpcHandler,
			server.WithServiceAddr(addr),
			server.WithSuite(tracing.NewServerSuite()),
			//server.WithMiddleware(middleware.PrintKitexRequestResponse),
		)
		e.grpcServer = svr
		klog.Infof("executor try start serve, protoc:grpc, port:%v", e.serveConf.Port)
		if err := svr.Run(); err != nil {
			panic(err)
		}
		break
	case discovery.ProtocTypeHttp:
		//todo
		break
	default:
		break
	}
}

func (e *Executor) register() error {
	ins := new(discovery.ServiceInstance)
	ins.ServiceServeConf = *e.serveConf
	ins.Meta = make(map[string]string, 1)
	ins.Meta[constance.ExecutorTagFieldName] = util.EncodeTag(e.tags)
	ins.Meta[constance.HealthCheckPortFieldName] = e.extraConf[myConstance.ConsulHealthCheckPortExtraConfKeyName]
	ins.ServiceName = constance.ExecutorServiceName
	ins.InstanceId = e.instanceID

	klog.Infof("executor try register service: %+v", ins)
	return e.discoveryClient.Register(ins)
}

func (e *Executor) Start() error {
	var (
		err error
	)

	//开启提供executor服务
	go e.startServe()

	e.executeService.Start()

	//注册到用户自己指定的中间件中
	if err = e.register(); err != nil {
		_ = e.grpcServer.Stop()
		e.executeService.Stop()
		return err
	}

	return nil
}

func (e *Executor) Stop() {
	_ = e.grpcServer.Stop()
	_ = e.discoveryClient.DiscoverServices(e.instanceID)
	e.executeService.Stop()
}