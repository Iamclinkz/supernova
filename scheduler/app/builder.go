package app

import (
	"errors"
	"strconv"
	"supernova/pkg/conf"
	"supernova/pkg/discovery"
	"supernova/scheduler/dal"
	"supernova/scheduler/operator/schedule_operator"
)

type SchedulerBuilder struct {
	scheduleOperator     schedule_operator.Operator
	discoveryClient      discovery.ExecutorDiscoveryClient
	schedulerWorkerCount int
	err                  error
}

func NewSchedulerBuilder() *SchedulerBuilder {
	return &SchedulerBuilder{}
}

func (b *SchedulerBuilder) WithConsulDiscovery(config *conf.ConsulConf) *SchedulerBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(discovery.TypeConsul,
		discovery.NewConsulMiddlewareConfig(config.Host, strconv.Itoa(config.Port)), nil)
	if err != nil && b.err == nil {
		b.err = err
	} else {
		b.discoveryClient = discoveryClient
	}

	return b
}

func (b *SchedulerBuilder) WithMysqlStore(config *conf.MysqlConf) *SchedulerBuilder {
	mysqlCli, err := dal.NewMysqlClient(config)
	if err != nil && b.err == nil {
		b.err = err
	} else {
		b.scheduleOperator, err = schedule_operator.NewMysqlScheduleOperator(mysqlCli)
		if err != nil && b.err == nil {
			b.err = err
		}
	}

	return b
}

func (b *SchedulerBuilder) WithSchedulerWorkerCount(count int) *SchedulerBuilder {
	if count <= 0 || count > 32 {
		b.err = errors.New("scheduler worker should be range in [1,32]")
	} else {
		b.schedulerWorkerCount = count
	}

	return b
}

func (b *SchedulerBuilder) Build() (*Scheduler, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.scheduleOperator == nil {
		return nil, errors.New("no select db")
	}
	if b.discoveryClient == nil {
		return nil, errors.New("no select discovery")
	}
	if b.schedulerWorkerCount == 0 {
		b.schedulerWorkerCount = 512
	}

	return genScheduler(b.scheduleOperator, b.discoveryClient, b.schedulerWorkerCount)
}
