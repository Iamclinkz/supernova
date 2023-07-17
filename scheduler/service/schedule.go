package service

import (
	"context"
	"errors"
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/pkg/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/util"
	"time"
)

type ScheduleService struct {
	shutdownCh            chan struct{}
	statisticsService     *StatisticsService
	jobService            *JobService
	triggerService        *TriggerService
	onFireService         *OnFireService
	executorSelectService *ExecutorRouteService
	//这里不用cron.Cron，因为触发时间确定且只触发一次且距离触发时间较短，所以用时间轮定时器了
	timeWheel *util.TimeWheel
}

func NewScheduleService(statisticsService *StatisticsService,
	jobService *JobService, triggerService *TriggerService, onFireService *OnFireService,
	executorSelectService *ExecutorRouteService) *ScheduleService {
	tw, _ := util.NewTimeWheel(time.Millisecond*20, 512, util.TickSafeMode())
	return &ScheduleService{
		shutdownCh:            make(chan struct{}),
		statisticsService:     statisticsService,
		jobService:            jobService,
		triggerService:        triggerService,
		onFireService:         onFireService,
		executorSelectService: executorSelectService,
		timeWheel:             tw,
	}
}

func (s *ScheduleService) Schedule() {
	ticker := time.NewTicker(s.statisticsService.GetScheduleInterval())

	for {
		select {
		case <-s.shutdownCh:
			ticker.Stop()
			return
		case <-ticker.C:
			onFireLogs, err := s.triggerService.fetchUpdateMarkTrigger(context.TODO(), true)
			if err != nil {
				klog.Errorf("Schedule Error:%v", err)
			} else {
				now := time.Now()
				for _, onFireLog := range onFireLogs {
					s.timeWheel.Add(onFireLog.ShouldFireAt.Sub(now), func() {
						if fireErr := s.fire(onFireLog); fireErr != nil {
							klog.Error("fire error: ", fireErr)
						}
					})
				}
			}
			ticker.Reset(s.statisticsService.GetScheduleInterval())
		}
	}
}

func (s *ScheduleService) Stop() {
	//同步等待
	s.shutdownCh <- struct{}{}
	s.timeWheel.Stop()

	//todo 怎么优雅退出？
}

func (s *ScheduleService) fire(onFireLog *model.OnFireLog) error {
	job, err := s.jobService.FetchJobFromID(context.TODO(), onFireLog.JobID)
	if err != nil {
		if err == schedule_operator.ErrNotFound {
			//job找不到了，只可能说明被删了。这种情况下，本trigger应该被删掉，同时onFireLog中的内容也没啥用了，也删掉
			_ = s.onFireService.DeleteOnFireLogFromID(context.TODO(), onFireLog.ID)
			_ = s.triggerService.DeleteTriggerFromID(context.TODO(), onFireLog.TriggerID)
			return errors.New("job was deleted")
		} else {
			//其他情况，可能是网络不通？先不操作了
			return errors.New("fetch db failed:" + err.Error())
		}
	}

	executor, err := s.executorSelectService.ChooseJobExecutor(job)
	if err != nil {
		return err
	}

	//更新db中的onFireLog
	onFireLog.ExecutorInstance = executor.Executor.InstanceId
	onFireLog.Status = constance.OnFireStatusExecuting
	if err = s.jobService.UpdateOnFireLogExecutorStatus(context.TODO(), onFireLog); err != nil {
		return err
	}

	//todo 这里再想一下，并且要插入一下日志
	_, err = executor.Operator.RunJob(nil, time.Now().Sub(onFireLog.TimeoutAt))
	if err != nil {
		return err
	}

	return nil
}
