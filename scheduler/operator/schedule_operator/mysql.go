package schedule_operator

import (
	"context"
	"errors"
	"fmt"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/dal"
	"supernova/scheduler/model"
	"time"

	"gorm.io/gorm"
)

var _ Operator = (*MysqlOperator)(nil)

type MysqlOperator struct {
	db             *dal.MysqlClient
	emptyJob       *model.Job
	emptyTrigger   *model.Trigger
	emptyOnFireLog *model.OnFireLog
	emptyLock      *model.Lock
}

var (
	reduceRedoAtExpr = gorm.Expr("left_try_count - 1")
)

func (m *MysqlOperator) UpdateOnFireLogRedoAt(ctx context.Context, onFireLog *model.OnFireLog) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	result := tx.Model(onFireLog).
		Where("id = ?", onFireLog.ID).
		Where("status != ?", constance.OnFireStatusFinished).
		Where("updated_at = ?", onFireLog.UpdatedAt).
		Updates(map[string]interface{}{
			"redo_at": onFireLog.RedoAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *MysqlOperator) FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.OnFireLog, error) {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	var logs []*model.OnFireLog

	err := tx.Select("id, trigger_id, job_id, executor_instance, param,redo_at,updated_at").
		Where("redo_at > ? AND redo_at < ?", noEarlyThan, noLaterThan).
		Where("status != ? ", constance.OnFireStatusFinished).
		Where("left_try_count > 0").
		Limit(maxCount).Find(&logs).Error
	return logs, err
}

func (m *MysqlOperator) UpdateOnFireLogStop(ctx context.Context, onFireLog *model.OnFireLog, msg string) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	// 更新满足条件的记录
	return tx.Model(onFireLog).
		Where("id = ? AND status != ?", onFireLog.ID, constance.OnFireStatusFinished).
		Updates(map[string]interface{}{
			"status":  constance.OnFireStatusFinished,
			"result":  msg,
			"redo_at": util.VeryLateTime(),
		}).Error
}

func (m *MysqlOperator) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	// 更新满足条件的记录
	return tx.Model(&model.OnFireLog{}).
		Where("id = ? AND status != ?", onFireLogID, constance.OnFireStatusFinished).
		Updates(map[string]interface{}{
			"success": true,
			"status":  constance.OnFireStatusFinished,
			"result":  result,
			"redo_at": util.VeryLateTime(),
		}).Error
}

func (m *MysqlOperator) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	// 更新满足条件的记录
	return tx.Model(&model.OnFireLog{}).
		Where("id = ? AND status != ? AND left_try_count > 0", onFireLogID, constance.OnFireStatusFinished).
		Updates(map[string]interface{}{
			"left_try_count": reduceRedoAtExpr,
			"result":         errorMsg,
		}).Error
}

func (m *MysqlOperator) FindOnFireLogByJobID(ctx context.Context, jobID uint) ([]*model.OnFireLog, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MysqlOperator) FetchOnFireLogByID(ctx context.Context, jobID uint) (*model.OnFireLog, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MysqlOperator) FindOnFireLogByTriggerID(ctx context.Context, triggerID uint) ([]*model.OnFireLog, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MysqlOperator) InsertJobs(ctx context.Context, jobs []*model.Job) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	err := tx.Create(jobs).Error
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	return nil
}

func (m *MysqlOperator) FindJobByName(ctx context.Context, jobName string) (*model.Job, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	job := new(model.Job)
	err := db.Where("job_name = ?", jobName).Find(job).Error

	if job.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (m *MysqlOperator) InsertTriggers(ctx context.Context, triggers []*model.Trigger) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	err := tx.Create(triggers).Error
	if err != nil {
		return fmt.Errorf("failed to insert trigger: %w", err)
	}

	return nil
}

func (m *MysqlOperator) FindTriggerByName(ctx context.Context, triggerName string) (*model.Trigger, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	trigger := new(model.Trigger)
	err := db.Where("trigger_name = ?", triggerName).Find(trigger).Error

	if trigger.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (m *MysqlOperator) Lock(ctx context.Context, lockName string) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	var lock model.Lock
	return tx.Raw("SELECT * FROM t_lock WHERE lock_name = ? FOR UPDATE", lockName).Scan(&lock).Error
}

func NewMysqlScheduleOperator(cli *dal.MysqlClient) (*MysqlOperator, error) {
	ret := &MysqlOperator{
		db:             cli,
		emptyOnFireLog: &model.OnFireLog{},
		emptyTrigger:   &model.Trigger{},
		emptyJob:       &model.Job{},
		emptyLock:      &model.Lock{},
	}

	//todo 测试用，记得删掉
	cli.DB().Migrator().DropTable(ret.emptyJob)
	cli.DB().Migrator().DropTable(ret.emptyTrigger)
	cli.DB().Migrator().DropTable(ret.emptyOnFireLog)
	cli.DB().Migrator().DropTable(ret.emptyLock)

	//todo 建表语句，实际上可以放到.sql文件中
	if err := cli.DB().AutoMigrate(ret.emptyJob); err != nil {
		return nil, err
	}
	if err := cli.DB().AutoMigrate(ret.emptyTrigger); err != nil {
		return nil, err
	}
	if err := cli.DB().AutoMigrate(ret.emptyOnFireLog); err != nil {
		return nil, err
	}
	if err := cli.DB().AutoMigrate(ret.emptyLock); err != nil {
		return nil, err
	}

	//初始化锁结构
	cli.DB().Create(&model.Lock{LockName: constance.FetchUpdateMarkTriggerLockName})
	return ret, nil
}

func (m *MysqlOperator) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	if result := tx.Unscoped().Where("id = ?", triggerID).Delete(m.emptyTrigger); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return fmt.Errorf("no triggerID:%v", triggerID)
	}

	return nil
}

func (m *MysqlOperator) DeleteOnFireLogFromID(ctx context.Context, onFireLogID uint) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	if err := tx.Unscoped().Where("id = ?", onFireLogID).Delete(m.emptyOnFireLog).Error; err != nil {
		return fmt.Errorf("failed to delete on fire log: %w", err)
	}

	return nil
}

const transactionKey string = "transaction"

func (m *MysqlOperator) OnTxStart(ctx context.Context) (context.Context, error) {
	//如果已经有事务了
	_, ok := ctx.Value(transactionKey).(*gorm.DB)
	if ok {
		return ctx, nil
	}

	tx := m.db.DB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	return context.WithValue(ctx, transactionKey, tx), nil
}

func (m *MysqlOperator) FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	var triggers []*model.Trigger
	err := db.Where("trigger_next_time BETWEEN ? AND ?", noEarlyThan, noLaterThan).
		Where("status = ?", constance.TriggerStatusNormal).Limit(maxCount).Find(&triggers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch triggers: %w", err)
	}

	return triggers, nil
}

func (m *MysqlOperator) UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	err := db.Save(triggers).Error
	if err != nil {
		return fmt.Errorf("failed to update triggers: %w", err)
	}

	return nil
}

func (m *MysqlOperator) InsertOnFires(ctx context.Context, onFires []*model.OnFireLog) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	err := db.Create(onFires).Error
	if err != nil {
		return fmt.Errorf("failed to insert on_fire record: %w", err)
	}

	return nil
}

func (m *MysqlOperator) OnTxFinish(ctx context.Context) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		return errors.New("no transaction in context")
	}

	if err := db.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *MysqlOperator) FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	job := new(model.Job)
	err := db.Where("id = ?", jobID).Find(job).Error

	if job.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (m *MysqlOperator) OnTxFail(ctx context.Context) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		return errors.New("no transaction in context")
	}

	return db.Rollback().Error
}

func (m *MysqlOperator) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	//为了避免同一个任务在多个Executor实例处重复执行，这里需要更新Executor的InstanceID。
	//测试用例跑了1000次发现个问题，就是假设SchedulerA取了TriggerA，并且插入了OnFireLogA，但是SchedulerA内部任务太多，
	//轮到OnFireLogA执行的时候，已经到了OnFireLogA中指定的RedoAt之后。这种情况下，另一个ShedulerB（或者该Scheduler自己）
	//拿到了过期的OnFireLog执行的话，发现Executor的InstanceID为空，所以自己按照用户指定的路由策略，分配一个Executor。
	//但是如果用户指定的是例如随机策略，那么很可能SchedulerA和SchedulerB就把同一个任务分配到不同的Executor上了。
	//但两个不同的Executor之间没有防重，所以同一个Trigger的一次触发被执行了两次。解决这个问题的方法是在任务执行之前，
	//通过redo_at字段再通过数据库看看，是不是我这次执行。如果失败，那么说明另一处已经要执行了。自己不需要不执行。
	result := db.Model(onFireLog).
		Where("id = ?", onFireLog.ID).
		Where("updated_at = ?", onFireLog.UpdatedAt).
		Updates(map[string]interface{}{
			"executor_instance": onFireLog.ExecutorInstance,
			"status":            onFireLog.Status,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *MysqlOperator) InsertJob(ctx context.Context, job *model.Job) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	err := db.Create(job).Error
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	return nil
}

func (m *MysqlOperator) InsertTrigger(ctx context.Context, trigger *model.Trigger) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	err := db.Create(trigger).Error
	if err != nil {
		return fmt.Errorf("failed to insert trigger: %w", err)
	}

	return nil
}

func (m *MysqlOperator) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	trigger := new(model.Trigger)
	err := db.Where("id = ?", triggerID).Find(trigger).Error

	if trigger.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (m *MysqlOperator) DeleteJobFromID(ctx context.Context, jobID uint) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	if result := tx.Unscoped().Where("id = ?", jobID).Delete(m.emptyJob); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return fmt.Errorf("no jobID:%v", jobID)
	}

	return nil
}

func (m *MysqlOperator) IsJobIDExist(ctx context.Context, jobID uint) (bool, error) {
	_, err := m.FetchJobFromID(ctx, jobID)

	if err == nil {
		return true, nil
	}

	if err == ErrNotFound {
		return false, nil
	}

	return false, err
}
