package app

import (
	"errors"
	"supernova/pkg/conf"
	"supernova/pkg/discovery"
	"supernova/scheduler/dal"
	"supernova/scheduler/operator/schedule_operator"
)

type SchedulerBuilder struct {
	scheduleOperator schedule_operator.Operator
	discoveryClient  discovery.Client
	err              error
}

func NewSchedulerBuilder() *SchedulerBuilder {
	return &SchedulerBuilder{}
}

func (b *SchedulerBuilder) WithConsulDiscovery(config *conf.ConsulConf) *SchedulerBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(discovery.TypeConsul, &discovery.Config{
		Host: config.Host,
		Port: config.Port,
	})
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

	return genScheduler(b.scheduleOperator, b.discoveryClient)
}
