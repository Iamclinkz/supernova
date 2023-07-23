package service

import (
	"context"
	"errors"
	"fmt"
	"supernova/pkg/util"
	"supernova/scheduler/config"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
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

func (s *TriggerService) DeleteTriggerFromID(triggerID uint) error {
	return s.scheduleOperator.DeleteTriggerFromID(context.TODO(), triggerID)
}

func (s *TriggerService) fetchUpdateMarkTrigger() ([]*model.OnFireLog, error) {
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
	if txCtx, err = s.scheduleOperator.OnTxStart(context.TODO()); err != nil {
		return nil, err
	}

	if err = s.scheduleOperator.Lock(txCtx, constance.FetchUpdateMarkTriggerLockName); err != nil {
		goto badEnd
	}

	//2.拿到位于[beginTriggerHandleTime,endTriggerHandleTime]之间的最近要执行的trigger
	fetchedTriggers, err = s.scheduleOperator.FetchRecentTriggers(txCtx,
		s.statisticsService.GetHandleTriggerMaxCount(), endTriggerHandleTime, beginTriggerHandleTime)
	if err != nil {
		goto badEnd
	}

	if len(fetchedTriggers) == 0 {
		goto emptyEnd
	}

	klog.Tracef("fetchUpdateMarkTrigger fetched triggers:%s", model.TriggersToString(fetchedTriggers))

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

			prepareFireErr := trigger.PrepareFire()

			if prepareFireErr != nil {
				//如果失败，那么设置trigger的状态为error，现在不执行，以后也不执行，直到用户手动处理
				trigger.Status = constance.TriggerStatusError
				trigger.TriggerNextTime = util.VeryLateTime()
				klog.Errorf("find error trigger:[%+v] when calling PrepareFire, err:%v", trigger, prepareFireErr)
				break
			}

			onFireTrigger := &model.OnFireLog{
				TriggerID:        trigger.ID,
				JobID:            trigger.JobID,
				Status:           constance.OnFireStatusWaiting,
				RetryCount:       trigger.FailRetryCount,
				ExecutorInstance: "",
				//下次重试时间 = 触发时间 + 用户指定执行最大时间 + 重试间隔 * 1
				RedoAt:            fireTime.Add(trigger.ExecuteTimeout).Add(config.JobRetryInterval),
				ShouldFireAt:      fireTime,
				ExecuteTimeout:    trigger.ExecuteTimeout,
				LeftRetryCount:    trigger.FailRetryCount + 1, //出错重试为1，表示出了两次错才不执行
				Param:             trigger.Param,
				FailRetryInterval: trigger.FailRetryInterval,
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

func (s *TriggerService) AddTrigger(trigger *model.Trigger) error {
	if err := s.ValidateTrigger(trigger); err != nil {
		return err
	}

	trigger.Status = constance.TriggerStatusNormal
	trigger.TriggerLastTime = util.VeryEarlyTime()
	if trigger.FailRetryInterval < time.Second {
		//重试间隔至少1s
		trigger.FailRetryInterval = time.Second
	}

	if err := s.scheduleOperator.InsertTrigger(context.TODO(), trigger); err != nil {
		return err
	}

	return nil
}

func (s *TriggerService) DeleteTrigger(triggerID uint) error {
	if err := s.scheduleOperator.DeleteTriggerFromID(context.TODO(), triggerID); err != nil {
		return err
	}

	return nil
}

func (s *TriggerService) ValidateTrigger(trigger *model.Trigger) error {
	// if trigger.Name == "" {
	// 	return errors.New("name cannot be empty")
	// }

	if !trigger.ScheduleType.Valid() {
		return errors.New("invalid ScheduleType")
	}

	if !trigger.MisfireStrategy.Valid() {
		return errors.New("invalid MisfireStrategyType")
	}

	if trigger.FailRetryCount < 0 {
		return errors.New("fail_retry_count must be greater than or equal to 0")
	}

	if trigger.ExecuteTimeout < time.Millisecond*20 {
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

func (s *TriggerService) FetchTriggerFromID(triggerID uint) (*model.Trigger, error) {
	trigger, err := s.scheduleOperator.FetchTriggerFromID(context.TODO(), triggerID)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (s *TriggerService) AddTriggers(triggers []*model.Trigger) error {
	for _, trigger := range triggers {
		if err := s.ValidateTrigger(trigger); err != nil {
			return fmt.Errorf("error trigger:%+v", trigger)
		}
		trigger.Status = constance.TriggerStatusNormal
		trigger.TriggerLastTime = util.VeryEarlyTime()
	}

	return s.scheduleOperator.InsertTriggers(context.TODO(), triggers)
}

func (s *TriggerService) FindTriggerByName(name string) (*model.Trigger, error) {
	return s.scheduleOperator.FindTriggerByName(context.TODO(), name)
}

func (s *TriggerService) fetchTimeoutAndRefreshOnFireLogs() ([]*model.OnFireLog, error) {
	var (
		now             = time.Now()
		beginHandleTime = now.Add(-s.statisticsService.GetHandleTriggerForwardDuration()) //最多到之前的多长时间
	)

	onFireLogs, err := s.scheduleOperator.FetchTimeoutOnFireLog(context.TODO(),
		s.statisticsService.GetHandleTimeoutOnFireLogMaxCount(), now, beginHandleTime)
	if err != nil {
		return nil, err
	}

	ret := make([]*model.OnFireLog, 0, len(onFireLogs))
	for _, onFireLog := range onFireLogs {
		if err = s.scheduleOperator.UpdateOnFireLogRedoAt(context.TODO(), onFireLog.ID, onFireLog.RedoAt); err == nil {
			//因为不考虑极端的情况下，失败的应该不多？且各个Scheduler动态调整捞取过期OnFireLog时间+捞取间隔较长，
			//所以这里用了乐观锁，取到过期的OnFireLog之后，尝试更新RedoAt字段。如果更新成功，则自己执行。更新失败则说明要不
			//成功了，要不让另一个进程抢先了，总之不是自己执行。
			ret = append(ret, onFireLog)
		}
	}

	if len(ret) != 0 {
		klog.Infof("fetchTimeoutAndRefreshOnFireLogs fetched timeout logs:%s", model.OnFireLogsToString(ret))
	} else {
		klog.Trace("fetchTimeoutAndRefreshOnFireLogs failed to fetch any timeout logs")
	}

	return ret, nil
}
