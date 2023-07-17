package app

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/service"
)

type Scheduler struct {
	//operator
	jobOperator schedule_operator.Operator

	//service
	scheduleService   *service.ScheduleService
	statisticsService *service.StatisticsService
	routerService     *service.ExecutorRouteService
	manageService     *service.ExecutorManageService
	jobService        *service.JobService
	triggerService    *service.TriggerService
	onFireService     *service.OnFireService
}

func newSchedulerInner(
	//operator
	jobOperator schedule_operator.Operator,
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
		//operator
		jobOperator: jobOperator,

		//service
		scheduleService:   scheduleService,
		statisticsService: statisticsService,
		routerService:     routerService,
		manageService:     manageService,
		jobService:        jobService,
		triggerService:    triggerService,
		onFireService:     onFireService,
	}
}

func (s *Scheduler) Start() {
	go s.scheduleService.Schedule()
	go s.manageService.HeartBeat()
	klog.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.scheduleService.Stop()
	s.manageService.Stop()
	klog.Info("Scheduler stopped")
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
