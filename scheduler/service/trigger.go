package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/pkg/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/util"
	"time"
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

func (s *TriggerService) fetchUpdateMarkTrigger(ctx context.Context, useTx bool) ([]*model.OnFireLog, error) {
	var (
		beginTriggerHandleTime = time.Now()
		endTriggerHandleTime   = beginTriggerHandleTime.Add(s.statisticsService.GetHandleTriggerDuration())
		err                    error
		txCtx                  = context.TODO()
		onFireTriggers         []*model.OnFireLog
	)

	if useTx {
		//1.开启事务，保证同时只能有一个实例拿到任务，且如果失败，则回滚
		if txCtx, err = s.scheduleOperator.OnTxStart(ctx); err != nil {
			return nil, err
		}
	}

	//2.拿到位于[beginTriggerHandleTime,endTriggerHandleTime]之间的最近要执行的trigger
	triggers, err := s.scheduleOperator.FetchRecentTriggers(txCtx, s.statisticsService.GetHandleTriggerMaxCount(), endTriggerHandleTime, beginTriggerHandleTime)
	if err != nil {
		goto badEnd
	}

	onFireTriggers = make([]*model.OnFireLog, 0, len(triggers))
	//3.依次让trigger更新自己，如果有问题的，则需要删除trigger
	for _, trigger := range triggers {
		//这里需要注意，如果一个trigger的触发时间很短，例如1s一次，而我们的HandleTriggerDuration较长，例如5s一次，
		//那么需要直接OnFire这个trigger 5次，并且更新trigger的nextFireTime到6s以后
		for true {
			fireTime := trigger.TriggerNextTime
			onFireLog := trigger.OnFire()

			if onFireLog != nil {
				//如果fire失败，那么设置trigger的状态为error，不处理后面的了
				trigger.Status = constance.TriggerStatusError
				trigger.TriggerNextTime = util.VeryLongTime()
				break
			}

			onFireTriggers = append(onFireTriggers, &model.OnFireLog{
				TriggerID:        trigger.ID,
				JobID:            trigger.JobID,
				Status:           constance.OnFireStatusWaiting,
				RetryCount:       trigger.FailRetryCount,
				ExecutorInstance: "",
				TimeoutAt:        fireTime.Add(trigger.TriggerTimeout),
				ShouldFireAt:     fireTime,
			})

			if trigger.TriggerNextTime.After(endTriggerHandleTime) {
				break
			}
		}
	}

	if err = s.scheduleOperator.InsertOnFires(txCtx, onFireTriggers); err != nil {
		goto badEnd
	}
	if err = s.scheduleOperator.UpdateTriggers(txCtx, triggers); err != nil {
		_ = s.scheduleOperator.OnTxFail(txCtx)
		goto badEnd
	}

	if useTx {
		if err = s.scheduleOperator.OnTxFinish(txCtx); err != nil {
			return nil, err
		}
	}

	return onFireTriggers, nil

badEnd:
	if useTx {
		if txFailErr := s.scheduleOperator.OnTxFail(txCtx); txFailErr != nil {
			return nil, errors.New(fmt.Sprintf("fetchUpdateMarkTrigger txFailErr:%v, originError:%v", txFailErr, err))
		}
	}
	return nil, err
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

	return nil
}

func (s *TriggerService) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	trigger, err := s.scheduleOperator.FetchTriggerFromID(ctx, triggerID)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}
