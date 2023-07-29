package app

import (
	"os"
	"os/signal"
	"supernova/executor/exporter"
	"supernova/executor/processor"
	"supernova/executor/service"
	"supernova/pkg/conf"
	"supernova/pkg/discovery"
	"syscall"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type Executor struct {
	//config
	instanceID     string
	tags           []string
	processor      map[string]processor.JobProcessor
	serveConf      *discovery.ExecutorServiceServeConf
	processorCount int
	extraConf      map[string]string

	//discovery
	discoveryClient discovery.ExecutorDiscoveryClient

	//service
	duplicateService  *service.DuplicateService
	executeService    *service.ExecuteService
	processorService  *service.ProcessorService
	statisticsService *service.StatisticsService

	serviceExporter exporter.Exporter
}

func newExecutorInner(
	instanceID string,
	tags []string,
	processor map[string]processor.JobProcessor,
	serveConf *discovery.ExecutorServiceServeConf,
	processorCount int,
	extraConf map[string]string,

	discoveryClient discovery.ExecutorDiscoveryClient,
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
		if err := ret.processorService.Register(p); err != nil {
			panic(err)
		}
	}

	ret.serviceExporter = exporter.NewExporter(executeService, statisticsService, serveConf)
	return ret
}

func (e *Executor) register() error {
	discoveryInstance := &discovery.ExecutorServiceInstance{
		InstanceId:               e.instanceID,
		ExecutorServiceServeConf: *e.serveConf,
		Tags:                     e.tags,
	}

	klog.Infof("executor try register service: %+v", discoveryInstance)
	return e.discoveryClient.Register(discoveryInstance)
}

func (e *Executor) Start() {
	var (
		err error
	)

	//创建execute worker，开始提供executor服务
	e.executeService.Start()
	go e.serviceExporter.StartServe()

	//注册到本服务到用户指定的服务发现中间件中
	if err = e.register(); err != nil {
		panic(err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	klog.Infof("executor started: %+v", e)
	s := <-signalCh
	klog.Infof("found signal:%v, start graceful stop", s)
	e.GracefulStop()
}

func (e *Executor) Stop() {
	_ = e.discoveryClient.DeRegister(e.instanceID)
	e.executeService.Stop()
	e.serviceExporter.Stop()
}

func (e *Executor) GracefulStop() {
	klog.Info("Executor start graceful stop")
	//1.从服务发现处注销自己。如果是consul之类的中间件，那么调用其取消注册api，新的scheduler下一次就不会发现自己了。
	//而如果是k8s，这里不需要取消注册，k8s滚动更新，如果决定干掉本pod，就不会导入流量给本pod了。所以不需要处理（from 常哥的指导）
	//这样做的好处是Executor和Scheduler之间的连接不需要断开。而如果Scheduler检测到来自Executor的连接断开，直接返回即可。
	err := e.discoveryClient.DeRegister(e.instanceID)
	if err != nil {
		klog.Errorf("fail to DeRegister executor service:%v", err)
	}

	//2.http/grpc不接受新连接
	switch e.serveConf.Protoc {
	case discovery.ProtocTypeGrpc:
		//这里想了一下，如果任务有重试，那么还得本Executor来做。
		//所以这里在Scheduler侧改成，如果发现Executor优雅退出，那么只发送重试任务，不发送新任务。
		//e.serviceExporter.GracefulStop()
		break
	default:
		//todo
		break
	}

	//3.通知statisticsService，下次Executor询问自己的健康情况时回复已经GracefulStop
	e.statisticsService.OnGracefulStop()

	//4.等待一个服务发现周期
	time.Sleep(conf.SchedulerMaxCheckHealthDuration + conf.DiscoveryMiddlewareCheckHeartBeatDuration)

	//5.等待所有任务处理结束
	for {
		time.Sleep(1 * time.Second)
		leftUnReplyRequest := e.statisticsService.GetUnReplyRequestCount()
		if leftUnReplyRequest == 0 {
			break
		}
		klog.Infof("Executor is waiting for leftUnReplyRequest, count: %v", leftUnReplyRequest)
	}

	//断开grpc连接
	e.Stop()

	klog.Info("Executor graceful stop success")
}
