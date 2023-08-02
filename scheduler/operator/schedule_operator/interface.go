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
	//-------------------------------onFire
	//增
	// 插入不带id的OnFireLog，插入数据库后从数据库中获取id并填充所有作为参数的onFireLog的ID字段
	InsertOnFires(ctx context.Context, onFire []*model.OnFireLog) error
	//改
	//改字段，需要使用Status和UpdateAt字段作为乐观锁的判断条件，即如果为Status为Finished，或者数据库中的UpdateAt和字段中的不一样，
	//则不应该更新，返回error。否则更新，并且更新UpdateAt字段到onFireLog参数中
	UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error
	UpdateOnFireLogRedoAt(ctx context.Context, onFireLog *model.OnFireLog) error
	//改状态，执行过一次UpdateOnFireLogSuccess，将状态迭代为success之后，后面再调用都不应该更新为fail
	//即需要使用Status作为乐观锁的判断条件。
	//更新失败，retry_count--
	UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error
	//更新成功，Status为Finished，Success为true，redo_at为很久以后
	UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error
	//停止，Status为Finished，redo_at为很久以后
	UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error
	//查
	//FindOnFireLogByJobID(ctx context.Context, jobID uint) ([]*model.OnFireLog, error)
	//FetchOnFireLogByID(ctx context.Context, jobID uint) (*model.OnFireLog, error)
	//FindOnFireLogByTriggerID(ctx context.Context, triggerID uint) ([]*model.OnFireLog, error)
	//查找过期的OnFireLog，要求至少返回id, trigger_id, job_id, executor_instance, param,redo_at,updated_at几个字段
	//并且筛选条件是过期时间（RedoAt）位于noEarlyThan和noLaterThan之间，并且Status不为Finished，并且left_try_count大于0。最大返回maxCount个。
	FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time, offset int) ([]*model.OnFireLog, error)

	//-------------------------------job
	//插入不带id的Job结构，插入数据库后从数据库中获取id并填充作为参数的Job的ID字段
	InsertJob(ctx context.Context, job *model.Job) error
	//插入不带id的Job切片结构，插入数据库后从数据库中获取id并填充所有作为参数的Job的ID字段
	InsertJobs(ctx context.Context, jobs []*model.Job) error
	//删
	DeleteJobFromID(ctx context.Context, jobID uint) error
	//改
	//查
	FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error)
	IsJobIDExist(ctx context.Context, jobID uint) (bool, error)
	//FindJobByName(ctx context.Context, jobName string) (*model.Job, error)

	//-------------------------------tx
	//仅限于支持事务的数据库实现。用于保护分布式锁。如果分布式锁本身有保护机制，或者数据库不支持事务，或者容忍任务丢失，那么可以不用实现
	OnTxStart(ctx context.Context) (context.Context, error)
	OnTxFail(ctx context.Context) error
	OnTxFinish(ctx context.Context) error

	//-------------------------------lock
	//分布式锁。如果不是单Scheduler，则必须实现。最好结合事务或者锁保护等机制
	Lock(ctx context.Context, lockName string) error

	//-------------------------------trigger
	//插入不带id的Trigger结构，插入数据库后从数据库中获取id并填充作为参数的Trigger的ID字段
	InsertTrigger(ctx context.Context, trigger *model.Trigger) error
	//插入不带id的Trigger切片结构，插入数据库后从数据库中获取id并所有填充作为参数的Trigger的ID字段
	InsertTriggers(ctx context.Context, triggers []*model.Trigger) error
	//删
	DeleteTriggerFromID(ctx context.Context, triggerID uint) error
	//改
	UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error
	//查
	FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error)
	FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error)
	FindTriggerByName(ctx context.Context, triggerName string) (*model.Trigger, error)
}
