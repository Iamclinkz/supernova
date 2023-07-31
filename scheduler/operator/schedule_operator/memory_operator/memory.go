package memory_operator

import (
	"context"
	"errors"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"sync"
	"time"

	"github.com/google/btree"
)

var _ schedule_operator.Operator = (*MemoryOperator)(nil)

type MemoryOperator struct {
	//trigger
	triggers         map[uint]*model.Trigger
	triggerTree      *btree.BTree //使用trigger.TriggerNextTime排序的b树。其实使用b+树更优，因为都是范围查询。但是先跑起来再调优吧
	triggerIDCounter uint
	triggerLock      sync.RWMutex

	//job
	jobs         map[uint]*model.Job
	jobLock      sync.RWMutex
	jobIDCounter uint

	//onFireLog
	onFireLogs         map[uint]*model.OnFireLog
	onFireLogTree      *btree.BTree //使用onFireLog.RedoAt排序的b树
	onFireLogIDCounter uint
	onFireLock         sync.RWMutex
}

func (m *MemoryOperator) InsertJobs(ctx context.Context, jobs []*model.Job) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	for _, job := range jobs {
		job.ID = m.jobIDCounter
		m.jobIDCounter++
		m.jobs[job.ID] = job
	}
	return nil
}

func (m *MemoryOperator) DeleteJobFromID(ctx context.Context, jobID uint) error {
	m.jobLock.Lock()
	defer m.jobLock.Unlock()

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

	newJob := &model.Job{}
	*newJob = *job
	return newJob, nil
}

func (m *MemoryOperator) IsJobIDExist(ctx context.Context, jobID uint) (bool, error) {
	m.jobLock.RLock()
	defer m.jobLock.RUnlock()

	_, ok := m.jobs[jobID]
	return ok, nil
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

func NewMemoryScheduleOperator() *MemoryOperator {
	return &MemoryOperator{
		triggers:      make(map[uint]*model.Trigger),
		triggerTree:   btree.New(5),
		jobs:          make(map[uint]*model.Job),
		onFireLogs:    make(map[uint]*model.OnFireLog),
		onFireLogTree: btree.New(5),
	}
}

func (m *MemoryOperator) InsertOnFires(ctx context.Context, onFire []*model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	for _, fire := range onFire {
		fire.ID = m.onFireLogIDCounter
		m.onFireLogIDCounter++
		m.onFireLogs[fire.ID] = fire
		m.onFireLogTree.ReplaceOrInsert(fire)
	}
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	dbOnFireLog, ok := m.onFireLogs[onFireLog.ID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if dbOnFireLog.Status == constance.OnFireStatusFinished || !dbOnFireLog.UpdatedAt.Equal(onFireLog.UpdatedAt) {
		return errors.New("optimistic lock error")
	}

	dbOnFireLog.ExecutorInstance = onFireLog.ExecutorInstance
	dbOnFireLog.Status = onFireLog.Status
	dbOnFireLog.UpdatedAt = time.Now()
	dbOnFireLog.TraceContext = onFireLog.TraceContext
	onFireLog.UpdatedAt = dbOnFireLog.UpdatedAt
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogRedoAt(ctx context.Context, onFireLog *model.OnFireLog) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	fire, ok := m.onFireLogs[onFireLog.ID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if fire.Status == constance.OnFireStatusFinished || !fire.UpdatedAt.Equal(onFireLog.UpdatedAt) {
		return errors.New("optimistic lock error")
	}

	fire.RedoAt = onFireLog.RedoAt
	fire.UpdatedAt = time.Now()
	onFireLog.UpdatedAt = fire.UpdatedAt
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
		return errors.New("optimistic lock error")
	}

	fire.LeftTryCount--
	fire.Result = errorMsg
	fire.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	fire, ok := m.onFireLogs[onFireLogID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if fire.Status == constance.OnFireStatusFinished {
		return errors.New("optimistic lock error")
	}

	fire.Success = true
	fire.Status = constance.OnFireStatusFinished
	fire.Result = result
	fire.RedoAt = util.VeryLateTime()
	fire.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryOperator) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	m.onFireLock.Lock()
	defer m.onFireLock.Unlock()

	fire, ok := m.onFireLogs[onFireLogID]
	if !ok {
		return schedule_operator.ErrNotFound
	}

	if fire.Status == constance.OnFireStatusFinished {
		return errors.New("optimistic lock error")
	}

	fire.Status = constance.OnFireStatusFinished
	fire.Result = msg
	fire.RedoAt = util.VeryLateTime()
	fire.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryOperator) InsertJob(ctx context.Context, job *model.Job) error {
	m.jobLock.Lock()
	defer m.jobLock.Unlock()

	m.jobs[job.ID] = job
	return nil
}

func (m *MemoryOperator) InsertTrigger(ctx context.Context, trigger *model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	m.triggers[m.triggerIDCounter] = trigger
	trigger.ID = m.triggerIDCounter
	m.triggerIDCounter++
	m.triggerTree.ReplaceOrInsert(trigger)
	return nil
}

func (m *MemoryOperator) InsertTriggers(ctx context.Context, triggers []*model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	for _, trigger := range triggers {
		m.triggers[trigger.ID] = trigger
	}
	return nil
}

func (m *MemoryOperator) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	delete(m.triggers, triggerID)
	return nil
}

func (m *MemoryOperator) UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error {
	m.triggerLock.Lock()
	defer m.triggerLock.Unlock()

	for _, trigger := range triggers {
		m.triggers[trigger.ID] = trigger
		// 更新 triggerTree 索引
		m.triggerTree.Delete(trigger)
		m.triggerTree.ReplaceOrInsert(trigger)
	}
	return nil
}

func (m *MemoryOperator) FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error) {
	m.triggerLock.RLock()
	defer m.triggerLock.RUnlock()

	var result []*model.Trigger
	m.triggerTree.AscendGreaterOrEqual(&model.Trigger{TriggerNextTime: noEarlyThan}, func(item btree.Item) bool {
		trigger := item.(*model.Trigger)
		if trigger.TriggerNextTime.Before(noLaterThan) && trigger.Status == constance.TriggerStatusNormal {
			result = append(result, trigger)
			return len(result) < maxCount
		}
		return false
	})
	return result, nil
}

func (m *MemoryOperator) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	m.triggerLock.RLock()
	defer m.triggerLock.RUnlock()

	trigger, ok := m.triggers[triggerID]
	if !ok {
		return nil, schedule_operator.ErrNotFound
	}

	newTrigger := &model.Trigger{}
	*newTrigger = *trigger
	return newTrigger, nil
}

func (m *MemoryOperator) FindTriggerByName(ctx context.Context, triggerName string) (*model.Trigger, error) {
	m.triggerLock.RLock()
	defer m.triggerLock.RUnlock()

	for _, trigger := range m.triggers {
		if trigger.Name == triggerName {
			newTrigger := &model.Trigger{}
			*newTrigger = *trigger
			return newTrigger, nil
		}
	}
	return nil, schedule_operator.ErrNotFound
}

func (m *MemoryOperator) FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.OnFireLog, error) {
	m.onFireLock.RLock()
	defer m.onFireLock.RUnlock()

	var result []*model.OnFireLog
	m.onFireLogTree.AscendGreaterOrEqual(&model.OnFireLog{RedoAt: noEarlyThan}, func(item btree.Item) bool {
		fire := item.(*model.OnFireLog)
		if fire.RedoAt.Before(noLaterThan) && fire.Status != constance.OnFireStatusFinished && fire.LeftTryCount > 0 {
			newFire := &model.OnFireLog{}
			*newFire = *fire
			result = append(result, newFire)
			return len(result) < maxCount
		}
		return false
	})
	return result, nil
}

var _ schedule_operator.Operator = (*MemoryOperator)(nil)
