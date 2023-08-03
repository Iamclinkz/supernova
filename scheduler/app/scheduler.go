package app

import (
	"context"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"
	"os/signal"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/pkg/session/trace"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
	"sync"
	"syscall"

	"github.com/cloudwego/kitex/pkg/klog"
)

type Scheduler struct {
	//config
	instanceID string //for debug
	standalone bool

	//openTelemetry
	oTelConfig     *trace.OTelConfig
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider

	//operator
	jobOperator schedule_operator.Operator

	discoveryClient discovery.DiscoverClient

	//service
	scheduleService   *service.ScheduleService
	statisticsService *service.StatisticsService
	routerService     *service.ExecutorRouteService
	manageService     *service.ExecutorManageService
	jobService        *service.JobService
	triggerService    *service.TriggerService
	onFireService     *service.OnFireService
	stopOnce          sync.Once
}

func newSchedulerInner(
	//config
	instanceID string,
	standAlone bool,

	//trace
	oTelConfig *trace.OTelConfig,

	//operator
	jobOperator schedule_operator.Operator,

	discoveryClient discovery.DiscoverClient,
	tracerProvider *sdktrace.TracerProvider,
	meterProvider *sdkmetric.MeterProvider,

	//service
	scheduleService *service.ScheduleService,
	statisticsService *service.StatisticsService,
	routerService *service.ExecutorRouteService,
	manageService *service.ExecutorManageService,
	jobService *service.JobService,
	triggerService *service.TriggerService,
	onFireService *service.OnFireService,
) *Scheduler {
	return &Scheduler{
		instanceID: instanceID,
		standalone: standAlone,

		oTelConfig:     oTelConfig,
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,

		//operator
		jobOperator: jobOperator,

		discoveryClient: discoveryClient,

		//service
		scheduleService:   scheduleService,
		statisticsService: statisticsService,
		routerService:     routerService,
		manageService:     manageService,
		jobService:        jobService,
		triggerService:    triggerService,
		onFireService:     onFireService,
		stopOnce:          sync.Once{},
	}
}

func (s *Scheduler) Start() {
	if err := s.discoveryClient.Register(&discovery.ServiceInstance{
		ServiceName: constance.SchedulerServiceName,
		InstanceId:  s.instanceID,
		ServiceServeConf: discovery.ServiceServeConf{
			Protoc: "Http",
			Host:   "", //暂时没啥用
			Port:   8080,
		},
		ExtraConfig: "",
	}); err != nil {
		panic(err)
	}
	go s.scheduleService.Schedule()
	go s.manageService.HeartBeat()
	go s.statisticsService.WatchScheduler()

	klog.Infof("Scheduler started: %+v", s)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signalCh
	klog.Infof("[%v]found signal:%v, start graceful stop", s.instanceID, sig)
	s.Stop()
}

func (s *Scheduler) Stop() {
	stopF := func() {
		s.scheduleService.Stop()
		s.manageService.Stop()
		s.statisticsService.Stop()
		if s.oTelConfig.EnableTrace {
			if err := s.tracerProvider.Shutdown(context.TODO()); err != nil {
				klog.Errorf("stop tracerProvider error:%v", err)
			}
		}
		if s.oTelConfig.EnableMetrics {
			if err := s.meterProvider.Shutdown(context.TODO()); err != nil {
				klog.Errorf("stop meterProvider error:%v", err)
			}
		}
		err := s.discoveryClient.DeRegister(s.instanceID)
		if err != nil {
			klog.Errorf("[%v] DeRegister error:%v", s.instanceID, err)
		}
		klog.Infof("[%v] stopped", s.instanceID)
	}
	s.stopOnce.Do(stopF)
}

func (s *Scheduler) GetJobOperator() schedule_operator.Operator {
	return s.jobOperator
}

func (s *Scheduler) GetScheduleService() *service.ScheduleService {
	return s.scheduleService
}

func (s *Scheduler) GetStatisticsService() *service.StatisticsService {
	return s.statisticsService
}

func (s *Scheduler) GetRouterService() *service.ExecutorRouteService {
	return s.routerService
}

func (s *Scheduler) GetManageService() *service.ExecutorManageService {
	return s.manageService
}

func (s *Scheduler) GetJobService() *service.JobService {
	return s.jobService
}

func (s *Scheduler) GetTriggerService() *service.TriggerService {
	return s.triggerService
}

func (s *Scheduler) GetOnFireService() *service.OnFireService {
	return s.onFireService
}
