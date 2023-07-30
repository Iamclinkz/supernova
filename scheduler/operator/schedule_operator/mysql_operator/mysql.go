package mysql_operator

import (
	"context"
	"errors"
	"fmt"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/dal"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/operator/schedule_operator/mysql_operator/dao"
	"time"

	"gorm.io/gorm"
)

var _ schedule_operator.Operator = (*MysqlOperator)(nil)

type MysqlOperator struct {
	db             *dal.MysqlClient
	emptyJob       *dao.Job
	emptyTrigger   *dao.Trigger
	emptyOnFireLog *dao.OnFireLog
	emptyLock      *dao.Lock
}

var (
	reduceRedoAtExpr = gorm.Expr("left_try_count - 1")
)

func NewMysqlScheduleOperator(cli *dal.MysqlClient) (*MysqlOperator, error) {
	ret := &MysqlOperator{
		db:             cli,
		emptyOnFireLog: &dao.OnFireLog{},
		emptyTrigger:   &dao.Trigger{},
		emptyJob:       &dao.Job{},
		emptyLock:      &dao.Lock{},
	}
	//todo 测试用，记得删掉
	cli.DB().Migrator().DropTable(ret.emptyTrigger)
	cli.DB().Migrator().DropTable(ret.emptyOnFireLog)
	if err := cli.DB().AutoMigrate(ret.emptyTrigger); err != nil {
		return nil, err
	}
	if err := cli.DB().AutoMigrate(ret.emptyOnFireLog); err != nil {
		return nil, err
	}

	cli.DB().Migrator().DropTable(ret.emptyJob)

	cli.DB().Migrator().DropTable(ret.emptyLock)

	//todo 建表语句，实际上可以放到.sql文件中
	if err := cli.DB().AutoMigrate(ret.emptyJob); err != nil {
		return nil, err
	}

	if err := cli.DB().AutoMigrate(ret.emptyLock); err != nil {
		return nil, err
	}

	//初始化锁结构
	cli.DB().Create(&dao.Lock{LockName: constance.FetchUpdateMarkTriggerLockName})
	return ret, nil
}

func (m *MysqlOperator) UpdateOnFireLogRedoAt(ctx context.Context, onFireLog *model.OnFireLog) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dOnFireLog, err := dao.FromModelOnFireLog(onFireLog)
	if err != nil {
		return err
	}

	result := db.Model(dOnFireLog).
		Where("id = ?", dOnFireLog.ID).
		Where("status != ?", constance.OnFireStatusFinished).
		Where("updated_at = ?", dOnFireLog.UpdatedAt).
		Updates(map[string]interface{}{
			"redo_at": dOnFireLog.RedoAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return schedule_operator.ErrNotFound
	}

	return nil
}

func (m *MysqlOperator) FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.OnFireLog, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	var dOnFireLogs []*dao.OnFireLog
	err := db.Model(&dao.OnFireLog{}).
		Select("id, trigger_id, job_id, executor_instance, param, redo_at, updated_at,trace_context").
		Where("redo_at BETWEEN ? AND ?", noEarlyThan, noLaterThan).
		Where("status != ?", constance.OnFireStatusFinished).
		Where("left_try_count > 0").
		Limit(maxCount).
		Find(&dOnFireLogs).Error

	if err != nil {
		return nil, err
	}

	onFireLogs := make([]*model.OnFireLog, len(dOnFireLogs))
	for i, dOnFireLog := range dOnFireLogs {
		onFireLog, err := dao.ToModelOnFireLog(dOnFireLog)
		if err != nil {
			return nil, err
		}
		onFireLogs[i] = onFireLog
	}

	return onFireLogs, nil
}

func (m *MysqlOperator) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	// 更新满足条件的记录
	return db.Model(&dao.OnFireLog{}).
		Where("id = ?", onFireLogID).
		Updates(map[string]interface{}{
			"status":  constance.OnFireStatusFinished,
			"result":  msg,
			"redo_at": util.VeryLateTime(),
		}).Error
}

func (m *MysqlOperator) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	// 更新满足条件的记录
	return db.Model(&dao.OnFireLog{}).
		Where("id = ? AND status != ?", onFireLogID, constance.OnFireStatusFinished).
		Updates(map[string]interface{}{
			"success": true,
			"status":  constance.OnFireStatusFinished,
			"result":  result,
			"redo_at": util.VeryLateTime(),
		}).Error
}

func (m *MysqlOperator) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	// 更新满足条件的记录
	return db.Model(&dao.OnFireLog{}).
		Where("id = ? AND status != ? AND left_try_count > 0", onFireLogID, constance.OnFireStatusFinished).
		Updates(map[string]interface{}{
			"left_try_count": reduceRedoAtExpr,
			"result":         errorMsg,
		}).Error
}

//func (m *MysqlOperator) FindOnFireLogByJobID(ctx context.Context, jobID uint) ([]*model.OnFireLog, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *MysqlOperator) FetchOnFireLogByID(ctx context.Context, jobID uint) (*model.OnFireLog, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *MysqlOperator) FindOnFireLogByTriggerID(ctx context.Context, triggerID uint) ([]*model.OnFireLog, error) {
//	//TODO implement me
//	panic("implement me")
//}

func (m *MysqlOperator) InsertJobs(ctx context.Context, jobs []*model.Job) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dJobs := make([]*dao.Job, len(jobs))
	for i, job := range jobs {
		dJob, err := dao.FromModelJob(job)
		if err != nil {
			return err
		}
		dJobs[i] = dJob
	}

	err := db.Create(dJobs).Error
	if err != nil {
		return fmt.Errorf("failed to insert jobs: %w", err)
	}

	for i, dJob := range dJobs {
		jobs[i].ID = dJob.ID
		jobs[i].UpdatedAt = dJob.UpdatedAt
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
		return nil, schedule_operator.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (m *MysqlOperator) InsertTriggers(ctx context.Context, triggers []*model.Trigger) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dTriggers := make([]*dao.Trigger, len(triggers))
	for i, trigger := range triggers {
		dTrigger, err := dao.FromModelTrigger(trigger)
		if err != nil {
			return err
		}
		dTriggers[i] = dTrigger
	}

	err := db.Create(dTriggers).Error
	if err != nil {
		return fmt.Errorf("failed to insert triggers: %w", err)
	}

	for i, dTrigger := range dTriggers {
		triggers[i].ID = dTrigger.ID
		triggers[i].UpdatedAt = dTrigger.UpdatedAt
	}

	return nil
}

func (m *MysqlOperator) FindTriggerByName(ctx context.Context, triggerName string) (*model.Trigger, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dTrigger := new(dao.Trigger)
	err := db.Where("name = ?", triggerName).Find(dTrigger).Error

	if dTrigger.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, schedule_operator.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	trigger, err := dao.ToModelTrigger(dTrigger)
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

	var lock dao.Lock
	return tx.Raw("SELECT * FROM t_lock WHERE lock_name = ? FOR UPDATE", lockName).Scan(&lock).Error
}

func (m *MysqlOperator) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	tx, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		tx = m.db.DB()
	}

	if result := tx.Unscoped().Where("id = ?", triggerID).Delete(&dao.Trigger{}); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return fmt.Errorf("no triggerID: %v", triggerID)
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

	var dTriggers []*dao.Trigger
	err := db.Where("trigger_next_time BETWEEN ? AND ?", noEarlyThan, noLaterThan).
		Where("status = ?", constance.TriggerStatusNormal).Limit(maxCount).Find(&dTriggers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch triggers: %w", err)
	}

	triggers := make([]*model.Trigger, len(dTriggers))
	for i, dTrigger := range dTriggers {
		trigger, err := dao.ToModelTrigger(dTrigger)
		if err != nil {
			return nil, err
		}
		triggers[i] = trigger
	}

	return triggers, nil
}

func (m *MysqlOperator) UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dTriggers := make([]*dao.Trigger, len(triggers))
	for i, trigger := range triggers {
		dTrigger, err := dao.FromModelTrigger(trigger)
		if err != nil {
			return err
		}
		dTriggers[i] = dTrigger
	}

	err := db.Save(dTriggers).Error
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

	dOnFires := make([]*dao.OnFireLog, len(onFires))
	for i, onFire := range onFires {
		dOnFire, err := dao.FromModelOnFireLog(onFire)
		if err != nil {
			return err
		}
		dOnFires[i] = dOnFire
	}

	err := db.Create(dOnFires).Error
	if err != nil {
		return fmt.Errorf("failed to insert on_fire record: %w", err)
	}

	for i, dOnFire := range dOnFires {
		onFires[i].ID = dOnFire.ID
		onFires[i].UpdatedAt = dOnFire.UpdatedAt
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

	dJob := new(dao.Job)
	err := db.Where("id = ?", jobID).Find(dJob).Error

	if dJob.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, schedule_operator.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	job, err := dao.ToModelJob(dJob)
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
	//轮到OnFireLogA执行的时候，已经到了OnFireLogA中指定的RedoAt之后。这种情况下，另一个SchedulerB（或者该Scheduler自己）
	//拿到了过期的OnFireLog执行的话，发现Executor的InstanceID为空，所以自己按照用户指定的路由策略，分配一个Executor。
	//但是如果用户指定的是例如随机策略，那么很可能SchedulerA和SchedulerB就把同一个任务分配到不同的Executor上了。
	//但两个不同的Executor之间没有防重，所以同一个Trigger的一次触发被执行了两次。解决这个问题的方法是在任务执行之前，
	//通过redo_at字段再通过数据库看看，是不是我这次执行。如果失败，那么说明另一处已经要执行了。自己不需要不执行。
	dOnFireLog, err := dao.FromModelOnFireLog(onFireLog)
	if err != nil {
		return err
	}

	result := db.Model(dOnFireLog).
		Where("id = ?", dOnFireLog.ID).
		Where("status != ?", constance.OnFireStatusFinished).
		Where("updated_at = ?", dOnFireLog.UpdatedAt).
		Updates(map[string]interface{}{
			"executor_instance": dOnFireLog.ExecutorInstance,
			"status":            dOnFireLog.Status,
			"trace_context":     dOnFireLog.TraceContext,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return schedule_operator.ErrNotFound
	}

	return nil
}

func (m *MysqlOperator) InsertJob(ctx context.Context, job *model.Job) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dJob, err := dao.FromModelJob(job)
	if err != nil {
		return err
	}

	err = db.Create(dJob).Error
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	job.ID = dJob.ID
	job.UpdatedAt = dJob.UpdatedAt
	return nil
}

func (m *MysqlOperator) InsertTrigger(ctx context.Context, trigger *model.Trigger) error {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dTrigger, err := dao.FromModelTrigger(trigger)
	if err != nil {
		return err
	}

	err = db.Create(dTrigger).Error
	if err != nil {
		return fmt.Errorf("failed to insert trigger: %w", err)
	}

	trigger.ID = dTrigger.ID
	trigger.UpdatedAt = dTrigger.UpdatedAt
	return nil
}

func (m *MysqlOperator) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	db, ok := ctx.Value(transactionKey).(*gorm.DB)
	if !ok {
		db = m.db.DB()
	}

	dTrigger := new(dao.Trigger)
	err := db.Where("id = ?", triggerID).Find(dTrigger).Error

	if dTrigger.ID == 0 || err == gorm.ErrRecordNotFound {
		return nil, schedule_operator.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	trigger, err := dao.ToModelTrigger(dTrigger)
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

	if result := tx.Unscoped().Where("id = ?", jobID).Delete(&dao.Job{}); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return fmt.Errorf("no jobID: %v", jobID)
	}

	return nil
}

func (m *MysqlOperator) IsJobIDExist(ctx context.Context, jobID uint) (bool, error) {
	_, err := m.FetchJobFromID(ctx, jobID)

	if err == nil {
		return true, nil
	}

	if err == schedule_operator.ErrNotFound {
		return false, nil
	}

	return false, err
}
