package app

import (
	"context"
	"os"
	"os/signal"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
	"sync"
	"syscall"

	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/cloudwego/kitex/pkg/klog"
)

type Scheduler struct {
	//config
	instanceID string //for debug

	//openTelemetry
	enableOTel     bool
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metric.MeterProvider

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

	//trace
	enableOTel bool,
	tracerProvider *sdktrace.TracerProvider,
	meterProvider *metric.MeterProvider,

	//operator
	jobOperator schedule_operator.Operator,

	discoveryClient discovery.DiscoverClient,

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

		enableOTel:     enableOTel,
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
		if s.enableOTel {
			if err := s.tracerProvider.Shutdown(context.TODO()); err != nil {
				klog.Errorf("stop tracerProvider error:%v", err)
			}
			if err := s.meterProvider.Shutdown(context.TODO()); err != nil {
				klog.Errorf("stop meterProvider error:%v", err)
			}
		}
		s.discoveryClient.DeRegister(s.instanceID)
		klog.Infof("%v stopped", s.instanceID)
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
