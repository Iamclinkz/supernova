package schedule_operator

import (
	"context"
	"errors"
	"supernova/scheduler/model"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type Operator interface {
	//-------------------------------tx
	OnTxStart(ctx context.Context) (context.Context, error)
	OnTxFail(ctx context.Context) error
	OnTxFinish(ctx context.Context) error
	Lock(ctx context.Context, lockName string) error

	//-------------------------------job
	//增
	InsertJob(ctx context.Context, job *model.Job) error
	InsertJobs(ctx context.Context, jobID []*model.Job) error
	//删
	DeleteJobFromID(ctx context.Context, jobID uint) error
	//改
	//查
	FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error)
	IsJobIDExist(ctx context.Context, jobID uint) (bool, error)
	FindJobByName(ctx context.Context, jobName string) (*model.Job, error)

	//-------------------------------trigger
	//增
	InsertTrigger(ctx context.Context, trigger *model.Trigger) error
	InsertTriggers(ctx context.Context, triggers []*model.Trigger) error
	//删
	DeleteTriggerFromID(ctx context.Context, triggerID uint) error
	//改
	UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error
	//查
	FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error)
	FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error)
	FindTriggerByName(ctx context.Context, triggerName string) (*model.Trigger, error)

	//-------------------------------onFire
	//增
	InsertOnFires(ctx context.Context, onFire []*model.OnFireLog) error
	//改
	UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error
	UpdateOnFireLogFail(ctx context.Context, onFireLogID uint) error
	UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error
	UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error
	UpdateOnFireLogRedoAt(ctx context.Context, onFireLogID uint, oldRedoAt time.Time) error

	//查
	FindOnFireLogByJobID(ctx context.Context, jobID uint) ([]*model.OnFireLog, error)
	FetchOnFireLogByID(ctx context.Context, jobID uint) (*model.OnFireLog, error)
	FindOnFireLogByTriggerID(ctx context.Context, triggerID uint) ([]*model.OnFireLog, error)
	FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.OnFireLog, error)
}
