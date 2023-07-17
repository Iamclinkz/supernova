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
	OnTxStart(ctx context.Context) (context.Context, error)
	OnTxFail(ctx context.Context) error
	FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error)
	UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error
	InsertOnFires(ctx context.Context, onFire []*model.OnFireLog) error
	OnTxFinish(ctx context.Context) error
	FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error)
	DeleteTriggerFromID(ctx context.Context, triggerID uint) error
	DeleteOnFireLogFromID(ctx context.Context, onFireLogID uint) error
	UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error
	InsertJob(ctx context.Context, job *model.Job) error
	InsertTrigger(ctx context.Context, trigger *model.Trigger) error
	FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error)
}
