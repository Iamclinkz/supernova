package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type TriggerService struct {
	jobService        *JobService
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService
	standalone        bool
	//for debug
	instanceID string
}

func NewTriggerService(scheduleOperator schedule_operator.Operator,
	statisticsService *StatisticsService, jobService *JobService, standalone bool, instanceID string) *TriggerService {
	return &TriggerService{
		scheduleOperator:  scheduleOperator,
		statisticsService: statisticsService,
		jobService:        jobService,
		standalone:        standalone,
		instanceID:        instanceID,
	}
}

func (s *TriggerService) DeleteTriggerFromID(triggerID uint) error {
	return s.scheduleOperator.DeleteTriggerFromID(context.TODO(), triggerID)
}

func (s *TriggerService) fetchUpdateMarkTrigger() ([]*model.OnFireLog, error) {
	var (
		begin                  = time.Now()
		beginTriggerHandleTime = begin.Add(-s.statisticsService.GetHandleTriggerForwardDuration())
		endTriggerHandleTime   = time.Now().Add(s.statisticsService.GetHandleTriggerDuration())
		err                    error
		txCtx                  = context.TODO()
		onFireLogs             []*model.OnFireLog
		fetchedTriggers        []*model.Trigger
		fetchLen               int
	)

	klog.Infof("[%v] try fetchUpdateMarkTrigger", s.instanceID)
	//1.开启事务，保证同时只能有一个实例拿到任务，且如果失败，则回滚
	if txCtx, err = s.scheduleOperator.OnTxStart(context.TODO()); err != nil {
		return nil, fmt.Errorf("OnTxStart error:%v", err)
	}

	if !s.standalone {
		//单机模式不需要加分布式锁
		if err = s.scheduleOperator.Lock(txCtx, constance.FetchUpdateMarkTriggerLockName); err != nil {
			err = fmt.Errorf("lock error:%v", err)
			goto badEnd
		}
	}

	//2.拿到位于[beginTriggerHandleTime,endTriggerHandleTime]之间的最近要执行的trigger
	fetchedTriggers, err = s.scheduleOperator.FetchRecentTriggers(txCtx,
		s.statisticsService.GetHandleTriggerMaxCount(), endTriggerHandleTime, beginTriggerHandleTime)
	if err != nil {
		err = fmt.Errorf("fetchRecentTriggers error:%v", err)
		goto badEnd
	}

	fetchLen = len(fetchedTriggers)
	if fetchLen == 0 {
		goto emptyEnd
	}

	//klog.Tracef("fetchUpdateMarkTrigger fetched triggers:%s", model.TriggersToString(fetchedTriggers))
	onFireLogs = make([]*model.OnFireLog, 0, fetchLen)
	//3.依次让trigger更新自己，如果有问题的，则需要删除trigger
	for _, trigger := range fetchedTriggers {
		//这里需要注意，如果一个trigger的触发时间很短，例如1s一次，而我们的HandleTriggerDuration较长，例如5s一次，
		//那么需要直接OnFire这个trigger 5次，并且更新trigger的nextFireTime到6s以后
		for {
			if trigger.TriggerNextTime.Before(begin) {
				//todo 这里需要处理misfire逻辑
				//这里赋值成fireTime，时间轮有容错。
				//todo:这里再想一下，暂时测试方便，改一下TriggerLastTime了，但是需要打散一下
				trigger.TriggerNextTime = time.Now().Add(util.GetRandomOffsetByNum(fetchLen))
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

			onFireLog := &model.OnFireLog{
				TriggerID:         trigger.ID,
				JobID:             trigger.JobID,
				Status:            constance.OnFireStatusWaiting,
				TryCount:          trigger.FailRetryCount,
				ExecutorInstance:  "",
				ShouldFireAt:      fireTime,
				LeftTryCount:      trigger.FailRetryCount,
				ExecuteTimeout:    trigger.ExecuteTimeout,
				Param:             trigger.Param,
				FailRetryInterval: trigger.FailRetryInterval,
				AtLeastOnce:       trigger.AtLeastOnce,
			}
			onFireLog.RedoAt = onFireLog.GetNextRedoAt()
			onFireLogs = append(onFireLogs, onFireLog)
			klog.Tracef("update on fire trigger:%+v", onFireLog)
			if trigger.TriggerNextTime.After(endTriggerHandleTime) {
				break
			}
		}
	}

	//不管怎么样都需要更新一下triggers，即使某个trigger发生了错误，也需要更新trigger的status为Error
	if err = s.scheduleOperator.UpdateTriggers(txCtx, fetchedTriggers); err != nil {
		err = fmt.Errorf("updateTriggers error:%v", err)
		goto badEnd
	}

	if len(onFireLogs) == 0 {
		klog.Warnf("fetch fetchedTriggers:%s, but none of them should be fired", fetchedTriggers)
		goto emptyEnd
	}

	if err = s.scheduleOperator.InsertOnFires(txCtx, onFireLogs); err != nil {
		err = fmt.Errorf("insertOnFires error:%v", err)
		goto badEnd
	}

	if !s.standalone {
		if err = s.scheduleOperator.UnLock(txCtx, constance.FetchUpdateMarkTriggerLockName); err != nil {
			err = fmt.Errorf("UnLock error:%v", err)
			goto badEnd
		}
	}

	if err = s.scheduleOperator.OnTxFinish(txCtx); err != nil {
		return nil, fmt.Errorf("onTxFinish error:%v", err)
	}

	klog.Tracef("fetchUpdateMarkTrigger fetched triggers, len:%v,use time:%v", len(onFireLogs), time.Since(begin))
	s.statisticsService.OnFetchNeedFireTriggers(len(onFireLogs))
	return onFireLogs, nil

badEnd:
	if !s.standalone {
		_ = s.scheduleOperator.UnLock(txCtx, constance.FetchUpdateMarkTriggerLockName)
	}
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
	_, err := s.jobService.FetchJobFromID(context.TODO(), trigger.JobID)
	if err != nil {
		return fmt.Errorf("fetch job error:%+v", err)
	}
	// if !jobExist {
	// 	return fmt.Errorf("no jobID:%v", trigger.JobID)
	// }

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

//func (s *TriggerService) FindTriggerByName(name string) (*model.Trigger, error) {
//	return s.scheduleOperator.FindTriggerByName(context.TODO(), name)
//}

func (s *TriggerService) fetchTimeoutAndRefreshOnFireLogs(closeCh chan struct{}, onFireLogCh chan *model.OnFireLog) {
	const (
		batchSize            = 50
		concurrentGoroutines = 4
	)

	for {
		select {
		case <-closeCh:
			return
		default:
			var (
				wg         sync.WaitGroup
				mu         sync.Mutex
				ret        []*model.OnFireLog
				failCount  atomic.Int32
				foundCount atomic.Int32
				startTime  = time.Now()
			)

			processTask := func(offset int) {
				defer wg.Done()
				//maxGetCount := s.statisticsService.GetHandleTimeoutOnFireLogMaxCount()

				beginHandleTime := startTime.Add(-s.statisticsService.GetHandleTriggerForwardDuration())

				onFireLogs, err := s.scheduleOperator.FetchTimeoutOnFireLog(context.TODO(), batchSize, startTime, beginHandleTime, offset)
				if err != nil {
					klog.Errorf("fetchTimeoutAndRefreshOnFireLogs error: %v", err)
					return
				} else {
					if len(onFireLogs) != 0 {
						klog.Tracef("fetch len:%v : %s", len(onFireLogs), model.OnFireLogsToString(onFireLogs))
					}
				}

				foundCount.Add(int32(len(onFireLogs)))

				//根据 RedoAt 字段排序
				sort.Slice(onFireLogs, func(i, j int) bool {
					return onFireLogs[i].RedoAt.Before(onFireLogs[j].RedoAt)
				})

				t2 := time.Now()
				for _, onFireLog := range onFireLogs {
					//打散任务触发时间
					onFireLog.ShouldFireAt = t2.Add(util.GetRandomOffsetByNum(1000))
					onFireLog.RedoAt = onFireLog.GetNextRedoAt()
					if err := s.scheduleOperator.UpdateOnFireLogRedoAt(context.TODO(), onFireLog); err == nil {
						mu.Lock()
						ret = append(ret, onFireLog)
						mu.Unlock()
					} else {
						failCount.Add(1)
						klog.Debugf("[%v] fetchTimeoutAndRefreshOnFireLogs fetch onFireLog [id-%v] error:%v", s.statisticsService.instanceID, onFireLog.ID, err)
					}
				}
			}

			for i := 0; i < concurrentGoroutines; i++ {
				wg.Add(1)
				go processTask(i * batchSize)
			}

			wg.Wait()

			var (
				found    = int(foundCount.Load())
				lost     = int(failCount.Load())
				got      = len(ret)
				failRate = float32(failCount.Load()) / float32(foundCount.Load())
			)

			if found == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			s.statisticsService.OnFindTimeoutOnFireLogs(found)
			s.statisticsService.OnHoldTimeoutOnFireLogFail(lost)
			s.statisticsService.OnHoldTimeoutOnFireLogSuccess(len(ret))
			s.statisticsService.UpdateLastTimeFetchTimeoutOnFireLogFailRate(failRate)

			//根据冲突概率调整下次拉取间隔
			sleepDuration := s.statisticsService.GetCheckTimeoutOnFireLogsInterval()

			for _, onFireLog := range ret {
				onFireLogCh <- onFireLog
			}

			//如果拉取的任务数量达到最大值且冲突概率低，则立即继续拉取
			if len(ret) == batchSize*concurrentGoroutines && failRate < 0.2 {
				continue
			}

			if got >= 0 {
				//获取成功一部分
				klog.Infof("[%v]fetchTimeoutAndRefreshOnFireLogs fetched timeout logs, len:%v", s.instanceID, len(ret))
			} else {
				//如果本轮没有获得任何一个OnFireLog
				if found > 0 {
					//查找到一些个过期的OnFireLog，但是自己一个都没获得到
					klog.Errorf("[%v]fetchTimeoutAndRefreshOnFireLogs find %v logs, but fetch nothing", s.instanceID, found)
				} else {
					//没有查找到过期的OnFireLog
					klog.Trace("[%v]fetchTimeoutAndRefreshOnFireLogs failed to fetch any timeout logs")
				}
			}

			klog.Infof("[%v]fetch timeout time:%v, got:%v, lost:%v", time.Since(startTime), s.instanceID, got, lost)

			//根据冲突概率调整的间隔休眠
			time.Sleep(sleepDuration)
		}
	}
}

func (s *TriggerService) fetchTimeoutAndRefreshOnFireLogsStandalone(closeCh chan struct{}, onFireLogCh chan *model.OnFireLog) {
	for {
		select {
		case <-closeCh:
			return
		default:
			var (
				now             = time.Now()
				beginHandleTime = now.Add(-s.statisticsService.GetHandleTriggerForwardDuration()) //最多到之前的多长时间
				batchSize       = 200
				wg              sync.WaitGroup
				mu              sync.Mutex
			)

			onFireLogs, err := s.scheduleOperator.FetchTimeoutOnFireLog(context.TODO(),
				s.statisticsService.GetHandleTimeoutOnFireLogMaxCount(), now, beginHandleTime, 0)
			if err != nil {
				klog.Errorf("[%v]fetchTimeoutAndRefreshOnFireLogsStandalone error in standalone mode")
				time.Sleep(1 * time.Second)
				continue
			}

			fetchLen := len(onFireLogs)
			if fetchLen == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			ret := make([]*model.OnFireLog, 0, fetchLen)

			processBatch := func(startIndex, endIndex int) {
				defer wg.Done()
				for i := startIndex; i < endIndex; i++ {
					onFireLog := onFireLogs[i]
					onFireLog.ShouldFireAt = time.Now().Add(util.GetRandomOffsetByNum(fetchLen))
					onFireLog.RedoAt = onFireLog.GetNextRedoAt()
					if err = s.scheduleOperator.UpdateOnFireLogRedoAt(context.TODO(), onFireLog); err == nil {
						//我们抢占这个待执行的过期OnFireLog成功
						mu.Lock()
						onFireLog.ShouldFireAt = time.Now()
						ret = append(ret, onFireLog)
						mu.Unlock()
					} else {
						klog.Debugf("[%v]fetchTimeoutAndRefreshOnFireLogs fetch onFireLog [id-%v] error:%v", onFireLog.ID, err)
					}
				}
			}

			n := len(onFireLogs)
			for i := 0; i < n; i += batchSize {
				startIndex := i
				endIndex := i + batchSize
				if endIndex > n {
					endIndex = n
				}
				wg.Add(1)
				go processBatch(startIndex, endIndex)
			}

			wg.Wait()

			for _, onFireLog := range ret {
				onFireLogCh <- onFireLog
			}

			var (
				gotCount  = len(ret)
				missCount = fetchLen - gotCount
			)

			s.statisticsService.OnFindTimeoutOnFireLogs(fetchLen)
			s.statisticsService.OnHoldTimeoutOnFireLogFail(missCount)
			s.statisticsService.OnHoldTimeoutOnFireLogSuccess(gotCount)
			if gotCount >= 0 {
				//获取成功一部分
				klog.Infof("[%v]fetchTimeoutAndRefreshOnFireLogs fetched timeout logs, len:%v", s.instanceID, len(ret))
			} else {
				//如果本轮没有获得任何一个OnFireLog
				if fetchLen > 0 {
					//查找到一些个过期的OnFireLog，但是自己一个都没获得到
					klog.Errorf("[%v]fetchTimeoutAndRefreshOnFireLogs find %v logs, but fetch nothing", s.instanceID, len(onFireLogs))
				} else {
					//没有查找到过期的OnFireLog
					klog.Trace("[%v]fetchTimeoutAndRefreshOnFireLogs failed to fetch any timeout logs", s.instanceID)
				}
			}
		}
	}
}
