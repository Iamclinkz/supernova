package memory_operator

import (
	"context"
	"errors"
	"strconv"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/operator/schedule_operator/memory_operator/dao"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/wangjia184/sortedset"
)

var _ schedule_operator.Operator = (*MemoryOperator)(nil)

type MemoryOperator struct {
	//trigger
	triggers map[uint]*dao.Trigger
	//使用trigger.TriggerNextTime排序的b树。value为dao.Trigger。其实使用b+树更优，因为都是范围查询。但是先跑起来再调优吧
	triggerTriggerNextTimeTree *sortedset.SortedSet
	triggerIDCounter           uint
	triggerLock                sync.RWMutex

	//job
	jobs         map[uint]*dao.Job
	jobLock      sync.RWMutex
	jobIDCounter uint

	//onFireLog
	onFireLogs map[uint]*dao.OnFireLog
	//使用RedoAt排序的b树。value为dao.RedoAt。
	onFireLogRedoAtTree *sortedset.SortedSet
	onFireLogIDCounter  uint
	onFireLock          sync.RWMutex

	//for test
	currentUnFinishOnFireLog atomic.Int64
	finishedOnFireLog        atomic.Int64
	fetchFailLog             atomic.Int64
	conflictCount            atomic.Int64
	failCount                atomic.Int64
}

func (m *MemoryOperator) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	//panic("implement me")
	return nil
}

func (m *MemoryOperator) InsertJob(ctx context.Context, job *model.Job) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	m.insertJobWithoutLock(job)
	return nil
}

func NewMemoryScheduleOperator() *MemoryOperator {
	ret := &MemoryOperator{
		triggers:                   make(map[uint]*dao.Trigger),
		triggerTriggerNextTimeTree: sortedset.New(),
		jobs:                       make(map[uint]*dao.Job),
		onFireLogs:                 make(map[uint]*dao.OnFireLog),
		onFireLogRedoAtTree:        sortedset.New(),
		currentUnFinishOnFireLog:   atomic.Int64{},
		finishedOnFireLog:          atomic.Int64{},
		fetchFailLog:               atomic.Int64{},
		conflictCount:              atomic.Int64{},
		failCount:                  atomic.Int64{},
	}

	go func() {
		for {
			time.Sleep(2 * time.Second)
			klog.Errorf("left onFireLog:%v, finished:%v, fetchFail:%v, conflictCount:%v, failCount:%v",
				ret.currentUnFinishOnFireLog.Load(),
				ret.finishedOnFireLog.Load(), ret.fetchFailLog.Load(), ret.conflictCount.Load(), ret.failCount.Load())
		}
	}()
	return ret
}

func (m *MemoryOperator) InsertJobs(ctx context.Context, jobs []*model.Job) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	for _, job := range jobs {
		m.insertJobWithoutLock(job)
	}
	return nil
}

func (m *MemoryOperator) insertJobWithoutLock(job *model.Job) {
	m.jobIDCounter++
	job.ID = m.jobIDCounter
	job.UpdatedAt = time.Now()
	m.jobs[job.ID] = dao.FromModelJob(job)
}

func (m *MemoryOperator) DeleteJobFromID(ctx context.Context, jobID uint) error {
	m.jobLock.Lock()
	defer m.jobLock.Unlock()

	if _, ok := m.jobs[jobID]; !ok {
		return schedule_operator.ErrNotFound
	}

	delete(m.jobs, jobID)
	return nil
}

func (m *MemoryOperator) FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error) {
	m.jobLock.RLock()
	defer m.jobLock.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, schedule_operator.ErrNotFound
	}

	return job.ToModelJob(), nil
}

func (m *MemoryOperator) IsJobIDExist(ctx context.Context, jobID uint) (bool, error) {
	m.jobLock.RLock()
	defer m.jobLock.RUnlock()

	_, ok := m.jobs[jobID]
	return ok, nil
}

func (m *MemoryOperator) insertJobOnFireWithoutLock(mOnFire *model.OnFireLog) {
	m.onFireLogIDCounter++
	mOnFire.ID = m.onFireLogIDCounter
	mOnFire.UpdatedAt = time.Now()

	dOnFire := dao.FromModelOnFireLog(mOnFire)
	m.onFireLogs[mOnFire.ID] = dOnFire
	score := float64(dOnFire.RedoAt.UnixNano())
	key := strconv.FormatUint(uint64(dOnFire.ID), 10)

	m.onFireLogRedoAtTree.AddOrUpdate(key, sortedset.SCORE(score), dOnFire)
	mOnFire.RedoAt = dOnFire.RedoAt
}

func (m *MemoryOperator) InsertOnFires(ctx context.Context, onFires []*model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	for _, fire := range onFires {
		m.insertJobOnFireWithoutLock(fire)
	}

	m.currentUnFinishOnFireLog.Add(int64(len(onFires)))
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	dbOnFireLog, ok := m.onFireLogs[onFireLog.ID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if dbOnFireLog.Status == constance.OnFireStatusFinished {
		return errors.New("already finished")
	} else if !dbOnFireLog.UpdatedAt.Equal(onFireLog.UpdatedAt) {
		return errors.New("updateAt not match")
	}

	dbOnFireLog.ExecutorInstance = onFireLog.ExecutorInstance
	dbOnFireLog.Status = onFireLog.Status
	dbOnFireLog.UpdatedAt = time.Now()
	dbOnFireLog.TraceContext = onFireLog.TraceContext
	onFireLog.UpdatedAt = dbOnFireLog.UpdatedAt
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogRedoAt(ctx context.Context, mOnFireLog *model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	df, ok := m.onFireLogs[mOnFireLog.ID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if df.Status == constance.OnFireStatusFinished {
		return errors.New("already finished")
	} else if !df.UpdatedAt.Equal(mOnFireLog.UpdatedAt) {
		//memory只适用于单机模式，这里不可能乐观锁失败，如果失败一定是代码错误
		m.conflictCount.Add(1)
		return errors.New("old data")
	}

	key := strconv.FormatUint(uint64(df.ID), 10)
	if m.onFireLogRedoAtTree.Remove(key) == nil {
		panic("")
	}

	df.RedoAt = mOnFireLog.RedoAt
	df.UpdatedAt = time.Now()
	score := float64(df.RedoAt.UnixNano())

	m.onFireLogRedoAtTree.AddOrUpdate(key, sortedset.SCORE(score), df)
	mOnFireLog.UpdatedAt = df.UpdatedAt
	mOnFireLog.RedoAt = df.RedoAt
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	fire, ok := m.onFireLogs[onFireLogID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if fire.Status == constance.OnFireStatusFinished {
		return errors.New("already finished")
	}

	fire.LeftTryCount--
	fire.Result = errorMsg
	fire.UpdatedAt = time.Now()
	m.failCount.Add(1)
	return nil
}

var (
	tmpMap = make(map[uint]struct{})
)

func (m *MemoryOperator) UpdateOnFireLogsSuccess(ctx context.Context, onFireLogs []struct {
	ID     uint
	Result string
}) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	for _, mf := range onFireLogs {
		fire, ok := m.onFireLogs[mf.ID]
		if !ok {
			return schedule_operator.ErrNotFound
		}

		if fire.Status == constance.OnFireStatusFinished {
			return errors.New("already finished")
		}

		key := strconv.FormatUint(uint64(fire.ID), 10)
		if m.onFireLogRedoAtTree.Remove(key) == nil {
			panic("")
		}

		fire.Success = true
		fire.Status = constance.OnFireStatusFinished
		fire.Result = mf.Result
		fire.RedoAt = util.VeryLateTime()
		fire.UpdatedAt = time.Now()

		m.currentUnFinishOnFireLog.Add(-1)
		m.finishedOnFireLog.Add(1)
	}

	return nil
}

//func (m *MemoryOperator) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
//	panic("")
//	m.onFireLock.Lock()
//	defer m.onFireLock.Unlock()
//
//	dOnFire, ok := m.onFireLogs[onFireLogID]
//	if !ok {
//		return schedule_operator.ErrNotFound
//	}
//
//	if dOnFire.Status == constance.OnFireStatusFinished {
//		return errors.New("already finished")
//	}
//
//	m.onFireLogRedoAtTree.Delete(dOnFire)
//	dOnFire.UpdatedAt = time.Now()
//	dOnFire.Status = constance.OnFireStatusFinished
//	dOnFire.Result = msg
//	dOnFire.RedoAt = util.VeryLateTime()
//	return nil
//}

//func (m *MemoryOperator) InsertJob(ctx context.Context, job *model.Job) error {
//	m.jobLock.Lock()
//	defer m.jobLock.Unlock()
//
//	m.insertJobWithoutLock(job)
//	return nil
//}

func (m *MemoryOperator) insertTriggerWithoutLock(trigger *model.Trigger) {
	m.triggerIDCounter++
	trigger.UpdatedAt = time.Now()
	trigger.ID = m.triggerIDCounter
	dt := dao.FromModelTrigger(trigger)
	m.triggers[m.triggerIDCounter] = dt
	score := float64(trigger.TriggerNextTime.UnixNano())
	key := strconv.FormatUint(uint64(trigger.ID), 10)
	m.triggerTriggerNextTimeTree.AddOrUpdate(key, sortedset.SCORE(score), dt)
}

func (m *MemoryOperator) InsertTrigger(ctx context.Context, trigger *model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()
	m.insertTriggerWithoutLock(trigger)
	return nil
}

func (m *MemoryOperator) InsertTriggers(ctx context.Context, triggers []*model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	for _, trigger := range triggers {
		m.insertTriggerWithoutLock(trigger)
	}
	return nil
}

func (m *MemoryOperator) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	panic("")
}

func (m *MemoryOperator) UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	for _, trigger := range triggers {
		dt, ok := m.triggers[trigger.ID]
		if !ok {
			panic("")
		}

		key := strconv.FormatUint(uint64(dt.ID), 10)
		if m.triggerTriggerNextTimeTree.Remove(key) == nil {
			panic("")
		}
		if trigger.Deleted {
			continue
		}

		trigger.UpdatedAt = time.Now()
		m.triggers[trigger.ID].Trigger = *trigger
		score := float64(trigger.TriggerNextTime.UnixNano())

		m.triggerTriggerNextTimeTree.AddOrUpdate(key, sortedset.SCORE(score), dt)
	}
	return nil
}

func (m *MemoryOperator) FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error) {
	m.triggerLock.RLock()
	defer m.triggerLock.RUnlock()

	ret := make([]*model.Trigger, 0, maxCount/2)
	startScore := float64(noEarlyThan.UnixNano())
	endScore := float64(noLaterThan.UnixNano())

	itemsInRange := m.triggerTriggerNextTimeTree.GetByScoreRange(sortedset.SCORE(startScore), sortedset.SCORE(endScore),
		&sortedset.GetByScoreRangeOptions{Limit: maxCount})

	for _, item := range itemsInRange {
		trigger := item.Value.(*dao.Trigger)
		if trigger.Status == constance.TriggerStatusNormal {
			ret = append(ret, trigger.ToModelOnFireLog())
		}
	}

	klog.Errorf("FetchRecentTriggers:%v", len(ret))
	return ret, nil
}

func (m *MemoryOperator) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	m.triggerLock.RLock()
	defer m.triggerLock.RUnlock()

	trigger, ok := m.triggers[triggerID]
	if !ok {
		return nil, schedule_operator.ErrNotFound
	}

	return trigger.ToModelOnFireLog(), nil
}

func (m *MemoryOperator) FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time, offset int) ([]*model.OnFireLog, error) {
	m.onFireLock.RLock()
	defer m.onFireLock.RUnlock()

	ret := make([]*model.OnFireLog, 0, maxCount/2)
	startScore := float64(noEarlyThan.UnixNano())
	endScore := float64(noLaterThan.UnixNano())

	itemsInRange := m.onFireLogRedoAtTree.GetByScoreRange(sortedset.SCORE(startScore), sortedset.SCORE(endScore),
		&sortedset.GetByScoreRangeOptions{Limit: maxCount})

	for _, item := range itemsInRange {
		fire := item.Value.(*dao.OnFireLog)
		if fire.Status != constance.OnFireStatusFinished && fire.LeftTryCount > 0 {
			ret = append(ret, fire.ToModelOnFireLog())
		}
	}

	m.fetchFailLog.Add(int64(len(ret)))
	return ret, nil
}

func (m *MemoryOperator) OnTxStart(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (m *MemoryOperator) OnTxFail(ctx context.Context) error {
	return nil
}

func (m *MemoryOperator) OnTxFinish(ctx context.Context) error {
	return nil
}

func (m *MemoryOperator) Lock(ctx context.Context, lockName string) error {
	return nil
}

func (m *MemoryOperator) UnLock(ctx context.Context, lockName string) error {
	return nil
}
