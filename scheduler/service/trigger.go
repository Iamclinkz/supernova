package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/util"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type TriggerService struct {
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService
}

func NewTriggerService(scheduleOperator schedule_operator.Operator, statisticsService *StatisticsService) *TriggerService {
	return &TriggerService{
		scheduleOperator:  scheduleOperator,
		statisticsService: statisticsService,
	}
}

func (s *TriggerService) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	return s.scheduleOperator.DeleteTriggerFromID(ctx, triggerID)
}

func (s *TriggerService) fetchUpdateMarkTrigger(ctx context.Context) ([]*model.OnFireLog, error) {
	var (
		now                    = time.Now()
		beginTriggerHandleTime = now.Add(-s.statisticsService.GetHandleTriggerForwardDuration())
		endTriggerHandleTime   = time.Now().Add(s.statisticsService.GetHandleTriggerDuration())
		err                    error
		txCtx                  = context.TODO()
		onFireTriggers         []*model.OnFireLog
		fetchedTriggers        []*model.Trigger
	)

	//1.开启事务，保证同时只能有一个实例拿到任务，且如果失败，则回滚
	if txCtx, err = s.scheduleOperator.OnTxStart(ctx); err != nil {
		return nil, err
	}

	if err = s.scheduleOperator.Lock(txCtx, constance.FetchUpdateMarkTriggerLockName); err != nil {
		goto badEnd
	}

	//2.拿到位于[beginTriggerHandleTime,endTriggerHandleTime]之间的最近要执行的trigger
	fetchedTriggers, err = s.scheduleOperator.FetchRecentTriggers(txCtx, s.statisticsService.GetHandleTriggerMaxCount(), endTriggerHandleTime, beginTriggerHandleTime)
	if err != nil {
		goto badEnd
	}

	if len(fetchedTriggers) == 0 {
		goto emptyEnd
	}

	klog.Trace(model.TriggersToString(fetchedTriggers))

	onFireTriggers = make([]*model.OnFireLog, 0, len(fetchedTriggers))
	//3.依次让trigger更新自己，如果有问题的，则需要删除trigger
	for _, trigger := range fetchedTriggers {
		//这里需要注意，如果一个trigger的触发时间很短，例如1s一次，而我们的HandleTriggerDuration较长，例如5s一次，
		//那么需要直接OnFire这个trigger 5次，并且更新trigger的nextFireTime到6s以后
		for {
			if trigger.TriggerNextTime.Before(now) {
				//todo 这里需要处理misfire逻辑
				//这里赋值成fireTime，时间轮有容错。
				//todo:这里再想一下，暂时测试方便，改一下TriggerLastTime了
				trigger.TriggerNextTime = now
			}
			fireTime := trigger.TriggerNextTime

			onFireErr := trigger.OnFire()

			if onFireErr != nil {
				//如果fire失败，那么设置trigger的状态为error，不处理后面的了
				trigger.Status = constance.TriggerStatusError
				trigger.TriggerNextTime = util.VeryLateTime()
				klog.Errorf("find error trigger:[%+v] when calling OnFire, err:%v", trigger, onFireErr)
				break
			}

			onFireTrigger := &model.OnFireLog{
				TriggerID:        trigger.ID,
				JobID:            trigger.JobID,
				Status:           constance.OnFireStatusWaiting,
				RetryCount:       trigger.FailRetryCount,
				ExecutorInstance: "",
				TimeoutAt:        fireTime.Add(trigger.TriggerTimeout),
				ShouldFireAt:     fireTime,
			}
			onFireTriggers = append(onFireTriggers, onFireTrigger)
			klog.Tracef("update on fire trigger:%+v", onFireTrigger)
			if trigger.TriggerNextTime.After(endTriggerHandleTime) {
				break
			}
		}
	}

	//不管怎么样都需要更新一下triggers，即使某个trigger发生了错误，也需要更新trigger的status为Error
	if err = s.scheduleOperator.UpdateTriggers(txCtx, fetchedTriggers); err != nil {
		goto badEnd
	}

	if len(onFireTriggers) == 0 {
		klog.Warnf("fetch fetchedTriggers:%s, but none of them should be fired", fetchedTriggers)
		goto emptyEnd
	}

	if err = s.scheduleOperator.InsertOnFires(txCtx, onFireTriggers); err != nil {
		goto badEnd
	}

	if err = s.scheduleOperator.OnTxFinish(txCtx); err != nil {
		return nil, err
	}

	return onFireTriggers, nil

badEnd:
	if txFailErr := s.scheduleOperator.OnTxFail(txCtx); txFailErr != nil {
		return nil, fmt.Errorf("fetchUpdateMarkTrigger txFailErr:%v, originError:%v", txFailErr, err)
	}
	return nil, err

emptyEnd:
	_ = s.scheduleOperator.OnTxFinish(txCtx)
	return []*model.OnFireLog{}, nil
}

func (s *TriggerService) AddTrigger(ctx context.Context, trigger *model.Trigger) error {
	if err := s.ValidateTrigger(trigger); err != nil {
		return err
	}

	if err := s.scheduleOperator.InsertTrigger(ctx, trigger); err != nil {
		return err
	}

	return nil
}

func (s *TriggerService) DeleteTrigger(ctx context.Context, triggerID uint) error {
	if err := s.scheduleOperator.DeleteTriggerFromID(ctx, triggerID); err != nil {
		return err
	}

	return nil
}

func (s *TriggerService) ValidateTrigger(trigger *model.Trigger) error {
	if trigger.Name == "" {
		return errors.New("name cannot be empty")
	}

	if !trigger.ScheduleType.Valid() {
		return errors.New("invalid ScheduleType")
	}

	if !trigger.MisfireStrategy.Valid() {
		return errors.New("invalid MisfireStrategyType")
	}

	if trigger.FailRetryCount < 0 {
		return errors.New("fail_retry_count must be greater than or equal to 0")
	}

	if trigger.TriggerTimeout < time.Millisecond*20 {
		return errors.New("trigger_timeout must be greater than 20ms")
	}

	//检查对应的job是否存在
	jobExist, err := s.scheduleOperator.IsJobIDExist(context.TODO(), trigger.JobID)
	if err != nil {
		return err
	}
	if !jobExist {
		return fmt.Errorf("no jobID:%v", trigger.JobID)
	}

	return nil
}

func (s *TriggerService) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	trigger, err := s.scheduleOperator.FetchTriggerFromID(ctx, triggerID)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (s *TriggerService) AddTriggers(ctx context.Context, triggers []*model.Trigger) error {
	for _, trigger := range triggers {
		if err := s.ValidateTrigger(trigger); err != nil {
			return fmt.Errorf("error trigger:%+v", trigger)
		}
	}

	return s.scheduleOperator.InsertTriggers(ctx, triggers)
}

func (s *TriggerService) FindTriggerByName(ctx context.Context, name string) (*model.Trigger, error) {
	return s.scheduleOperator.FindTriggerByName(ctx, name)
}
