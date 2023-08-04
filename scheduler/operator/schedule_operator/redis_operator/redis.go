package redis_operator

import (
	"context"
	"encoding/json"
	"strconv"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisOperator struct {
	redisClient *redis.Client
}

func NewRedisScheduleOperator(redisClient *redis.Client) *RedisOperator {
	exists, err := redisClient.Exists(context.TODO(), "jobIDCounter").Result()
	if err != nil {
		panic("")
	}

	if exists == 0 {
		err := redisClient.Set(context.TODO(), "jobIDCounter", 0, 0).Err()
		if err != nil {
			panic(err)
		}
	}

	return &RedisOperator{
		redisClient: redisClient,
	}
}

func (r *RedisOperator) InsertOnFires(ctx context.Context, onFires []*model.OnFireLog) error {
	onFireIDRange, err := r.redisClient.IncrBy(ctx, "onFireIDCounter", int64(len(onFires))).Result()
	if err != nil {
		return err
	}

	pipe := r.redisClient.Pipeline()
	for i, onFire := range onFires {
		onFire.ID = uint(onFireIDRange) - uint(len(onFires)) + uint(i) + 1
		onFireKey := "onFireLog:" + strconv.Itoa(int(onFire.ID))
		onFireValue, err := json.Marshal(onFire)
		if err != nil {
			return err
		}
		pipe.Set(ctx, onFireKey, onFireValue, 0)

		// 使用 ZADD 命令将 OnFireLog 插入到一个有序集合中，并使用 RedoAt 字段进行排序
		pipe.ZAdd(ctx, "onFireLogRedoAt", &redis.Z{
			Score:  float64(onFire.RedoAt.Unix()),
			Member: onFireKey,
		})
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisOperator) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	onFireKey := "onFireLog:" + strconv.Itoa(int(onFireLog.ID))

	err := r.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		onFireStr, err := tx.Get(ctx, onFireKey).Result()
		if err != nil {
			if err == redis.Nil {
				return schedule_operator.ErrNotFound
			}
			return err
		}

		var onFire model.OnFireLog
		err = json.Unmarshal([]byte(onFireStr), &onFire)
		if err != nil {
			return err
		}

		if !onFire.UpdatedAt.Equal(onFireLog.UpdatedAt) {
			return errors.New("old data")
		}

		onFire.ExecutorInstance = onFireLog.ExecutorInstance
		onFire.Status = onFireLog.Status
		onFire.UpdatedAt = time.Now()

		onFireValue, err := json.Marshal(onFire)
		if err != nil {
			return err
		}

		_, err = tx.Set(ctx, onFireKey, onFireValue, 0).Result()
		return err
	}, onFireKey)

	if err != nil {
		return err
	}

	return nil
}

func (r *RedisOperator) UpdateOnFireLogRedoAt(ctx context.Context, onFireLog *model.OnFireLog) error {
	onFireKey := "onFireLog:" + strconv.Itoa(int(onFireLog.ID))

	err := r.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		onFireStr, err := tx.Get(ctx, onFireKey).Result()
		if err != nil {
			if err == redis.Nil {
				return schedule_operator.ErrNotFound
			}
			return err
		}

		var onFire model.OnFireLog
		err = json.Unmarshal([]byte(onFireStr), &onFire)
		if err != nil {
			return err
		}

		// 检查乐观锁条件
		if !onFire.UpdatedAt.Equal(onFireLog.UpdatedAt) {
			return errors.New("old data")
		}

		pipe := tx.Pipeline()

		pipe.ZRem(ctx, "onFireLogRedoAt", onFireKey)

		onFire.RedoAt = onFireLog.RedoAt
		onFire.UpdatedAt = time.Now()

		onFireValue, err := json.Marshal(onFire)
		if err != nil {
			return err
		}

		pipe.Set(ctx, onFireKey, onFireValue, 0)

		pipe.ZAdd(ctx, "onFireLogRedoAt", &redis.Z{
			Score:  float64(onFire.RedoAt.Unix()),
			Member: onFireKey,
		})

		_, err = pipe.Exec(ctx)
		return err
	}, onFireKey)

	if err != nil {
		return err
	}

	return nil
}

// todo 有点问题
func (r *RedisOperator) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	onFireKey := "onFireLog:" + strconv.Itoa(int(onFireLogID))

	err := r.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		onFireStr, err := tx.Get(ctx, onFireKey).Result()
		if err != nil {
			if err == redis.Nil {
				return schedule_operator.ErrNotFound
			}
			return err
		}

		var onFire model.OnFireLog
		err = json.Unmarshal([]byte(onFireStr), &onFire)
		if err != nil {
			return err
		}

		// 检查条件
		if onFire.Status == constance.OnFireStatusFinished || onFire.LeftTryCount <= 0 {
			return errors.New("invalid status or no remaining attempts")
		}

		onFire.LeftTryCount--
		onFire.Result = errorMsg
		onFire.UpdatedAt = time.Now()

		onFireValue, err := json.Marshal(onFire)
		if err != nil {
			return err
		}

		_, err = tx.Set(ctx, onFireKey, onFireValue, 0).Result()
		return err
	}, onFireKey)

	if err != nil {
		return err
	}

	return nil
}

func (r *RedisOperator) UpdateOnFireLogsSuccess(ctx context.Context, onFireLogs []struct {
	ID     uint
	Result string
}) error {
	onFireKey := "onFireLog:" + strconv.Itoa(int(onFireLogID))

	err := r.redisClient.ZRem(ctx, "onFireLogRedoAt", onFireKey).Err()

	if err != nil {
		return err
	}

	return nil
}

func (r *RedisOperator) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisOperator) FetchTimeoutOnFireLog(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time, offset int) ([]*model.OnFireLog, error) {
	onFireLogKeys, err := r.redisClient.ZRangeByScore(ctx, "onFireLogRedoAt", &redis.ZRangeBy{
		Min:    strconv.FormatInt(noEarlyThan.Unix(), 10),
		Max:    strconv.FormatInt(noLaterThan.Unix(), 10),
		Offset: int64(offset),
		Count:  int64(maxCount),
	}).Result()

	if err != nil {
		return nil, err
	}

	if len(onFireLogKeys) == 0 {
		return []*model.OnFireLog{}, nil
	}

	onFireLogStrs, err := r.redisClient.MGet(ctx, onFireLogKeys...).Result()
	if err != nil {
		return nil, err
	}

	onFireLogs := make([]*model.OnFireLog, 0, len(onFireLogStrs))
	for _, onFireLogStr := range onFireLogStrs {
		if onFireLogStr == nil {
			continue
		}

		var onFireLog model.OnFireLog
		err = json.Unmarshal([]byte(onFireLogStr.(string)), &onFireLog)
		if err != nil {
			return nil, err
		}

		//todo 待修改
		if onFireLog.Status != constance.OnFireStatusFinished && onFireLog.LeftTryCount > 0 {
			onFireLogs = append(onFireLogs, &onFireLog)
		}
	}

	return onFireLogs, nil
}

func (r *RedisOperator) InsertJob(ctx context.Context, job *model.Job) error {
	jobID, err := r.redisClient.Incr(ctx, "jobIDCounter").Result()
	if err != nil {
		return err
	}

	job.ID = uint(jobID)
	jobKey := "job:" + strconv.Itoa(int(job.ID))
	jobValue, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.redisClient.Set(ctx, jobKey, jobValue, 0).Err()
}

func (r *RedisOperator) InsertJobs(ctx context.Context, jobs []*model.Job) error {
	jobIDRange, err := r.redisClient.IncrBy(ctx, "jobIDCounter", int64(len(jobs))).Result()
	if err != nil {
		return err
	}

	pipe := r.redisClient.Pipeline()
	for i, job := range jobs {
		job.ID = uint(jobIDRange) - uint(len(jobs)) + uint(i) + 1
		jobKey := "job:" + strconv.Itoa(int(job.ID))
		jobValue, err := json.Marshal(job)
		if err != nil {
			return err
		}
		pipe.Set(ctx, jobKey, jobValue, 0)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisOperator) DeleteJobFromID(ctx context.Context, jobID uint) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisOperator) FetchJobFromID(ctx context.Context, jobID uint) (*model.Job, error) {
	jobStr, err := r.redisClient.Get(ctx, "job:"+strconv.Itoa(int(jobID))).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, schedule_operator.ErrNotFound
		}
		return nil, err
	}

	var job model.Job
	err = json.Unmarshal([]byte(jobStr), &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *RedisOperator) IsJobIDExist(ctx context.Context, jobID uint) (bool, error) {
	count, err := r.redisClient.Exists(ctx, "job:"+strconv.Itoa(int(jobID))).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RedisOperator) OnTxStart(ctx context.Context) (context.Context, error) {
	return context.TODO(), nil
}

func (r *RedisOperator) OnTxFail(ctx context.Context) error {
	return nil
}

func (r *RedisOperator) OnTxFinish(ctx context.Context) error {
	return nil
}

const (
	lockExpiration = 10 * time.Second
)

func (r *RedisOperator) Lock(ctx context.Context, lockName string) error {
	for {
		ok, err := r.redisClient.SetNX(ctx, lockName, "1", lockExpiration).Result()
		if err != nil {
			return err
		}

		if ok {
			// 获取到锁，跳出循环
			break
		}

		// 等待一段时间后重试
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (r *RedisOperator) UnLock(ctx context.Context, lockName string) error {
	return r.redisClient.Del(ctx, lockName).Err()
}

func (r *RedisOperator) InsertTrigger(ctx context.Context, trigger *model.Trigger) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisOperator) InsertTriggers(ctx context.Context, triggers []*model.Trigger) error {
	triggerIDRange, err := r.redisClient.IncrBy(ctx, "triggerIDCounter", int64(len(triggers))).Result()
	if err != nil {
		return err
	}

	now := time.Now()
	pipe := r.redisClient.Pipeline()
	for i, trigger := range triggers {
		trigger.UpdatedAt = now
		trigger.ID = uint(triggerIDRange) - uint(len(triggers)) + uint(i) + 1
		triggerKey := "trigger:" + strconv.Itoa(int(trigger.ID))
		triggerValue, err := json.Marshal(trigger)
		if err != nil {
			return err
		}
		pipe.Set(ctx, triggerKey, triggerValue, 0)

		pipe.ZAdd(ctx, "triggerTriggerNextTime", &redis.Z{
			Score:  float64(trigger.TriggerNextTime.Unix()),
			Member: triggerKey,
		})
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisOperator) DeleteTriggerFromID(ctx context.Context, triggerID uint) error {
	//TODO implement me
	panic("implement me")
}

// todo 有错
func (r *RedisOperator) UpdateTriggers(ctx context.Context, triggers []*model.Trigger) error {
	for _, trigger := range triggers {
		triggerKey := "trigger:" + strconv.Itoa(int(trigger.ID))
		triggerValue, err := json.Marshal(trigger)
		if err != nil {
			return err
		}

		err = r.redisClient.Set(ctx, triggerKey, triggerValue, 0).Err()
		if err != nil {
			return err
		}

		err = r.redisClient.ZRem(ctx, "triggerTriggerNextTime", triggerKey).Err()
		if err != nil {
			return err
		}

		if !trigger.Deleted {
			err = r.redisClient.ZAdd(ctx, "triggerTriggerNextTime", &redis.Z{
				Score:  float64(trigger.TriggerNextTime.Unix()),
				Member: triggerKey,
			}).Err()

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RedisOperator) FetchRecentTriggers(ctx context.Context, maxCount int, noLaterThan, noEarlyThan time.Time) ([]*model.Trigger, error) {
	triggerKeys, err := r.redisClient.ZRangeByScore(ctx, "triggerTriggerNextTime", &redis.ZRangeBy{
		Min:   strconv.FormatInt(noEarlyThan.Unix(), 10),
		Max:   strconv.FormatInt(noLaterThan.Unix(), 10),
		Count: int64(maxCount),
	}).Result()

	if err != nil {
		return nil, err
	}

	// 如果 triggerKeys 为空，直接返回一个空列表
	if len(triggerKeys) == 0 {
		return []*model.Trigger{}, nil
	}

	triggerStrs, err := r.redisClient.MGet(ctx, triggerKeys...).Result()
	if err != nil {
		return nil, err
	}

	triggers := make([]*model.Trigger, len(triggerStrs))
	for i, triggerStr := range triggerStrs {
		if triggerStr == nil {
			return nil, errors.New("trigger not found")
		}

		var trigger model.Trigger
		err = json.Unmarshal([]byte(triggerStr.(string)), &trigger)
		if err != nil {
			return nil, err
		}
		triggers[i] = &trigger
	}

	return triggers, nil
}

func (r *RedisOperator) FetchTriggerFromID(ctx context.Context, triggerID uint) (*model.Trigger, error) {
	//TODO implement me
	panic("implement me")
}

var _ schedule_operator.Operator = (*RedisOperator)(nil)
