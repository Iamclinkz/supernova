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
	//tx
	OnTxStart(ctx context.Context) (context.Context, error)
	OnTxFail(ctx context.Context) error
	OnTxFinish(ctx context.Context) error

	//job
	FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error)
	DeleteJobFromID(ctx context.Context, jobID uint) error
	InsertJob(ctx context.Context, job *model.Job) error
	IsJobIDExist(ctx context.Context, jobID uint) (bool, error)

	//trigger
	FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error)
	UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error
	DeleteTriggerFromID(ctx context.Context, triggerID uint) error
	InsertTrigger(ctx context.Context, trigger *model.Trigger) error
	FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error)

	//onFire
	InsertOnFires(ctx context.Context, onFire []*model.OnFireLog) error
	DeleteOnFireLogFromID(ctx context.Context, onFireLogID uint) error
	UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error
}
