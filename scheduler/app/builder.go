package app

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"supernova/pkg/conf"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/pkg/session/trace"
	"supernova/scheduler/dal"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/operator/schedule_operator/mysql_operator"
)

type SchedulerBuilder struct {
	instanceID           string
	scheduleOperator     schedule_operator.Operator
	discoveryClient      discovery.ExecutorDiscoveryClient
	schedulerWorkerCount int
	oTelProvider         provider.OtelProvider
	err                  error
}

func NewSchedulerBuilder() *SchedulerBuilder {
	return &SchedulerBuilder{}
}

func (b *SchedulerBuilder) WithConsulDiscovery(consulHost, consulPort string) *SchedulerBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(
		discovery.TypeConsul,
		discovery.NewConsulMiddlewareConfig(consulHost, consulPort),
		nil)
	if err != nil && b.err == nil {
		b.err = err
	} else {
		b.discoveryClient = discoveryClient
	}

	return b
}

func (b *SchedulerBuilder) WithK8sDiscovery(namespace string) *SchedulerBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(discovery.TypeK8s,
		discovery.NewK8sMiddlewareConfig(namespace), nil)
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
		b.scheduleOperator, err = mysql_operator.NewMysqlScheduleOperator(mysqlCli)
		if err != nil && b.err == nil {
			b.err = err
		}
	}

	return b
}

func (b *SchedulerBuilder) WithMemoryStore() *SchedulerBuilder {
	b.scheduleOperator = schedule_operator.NewMemoryScheduleOperator()
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

func (b *SchedulerBuilder) WithInstanceID(instanceID string) *SchedulerBuilder {
	if instanceID == "" {
		b.err = errors.New("empty instanceID")
	} else {
		b.instanceID = instanceID
	}

	return b
}

func (b *SchedulerBuilder) WithTraceProvider(instrumentConf *conf.OTelConf) *SchedulerBuilder {
	b.oTelProvider = trace.InitProvider(constance.SchedulerServiceName, instrumentConf)
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
	if b.instanceID == "" {
		b.instanceID = fmt.Sprintf("Scheduler-%v", uuid.New())
	}

	return genScheduler(b.instanceID, b.oTelProvider == nil, b.oTelProvider, b.scheduleOperator, b.discoveryClient, b.schedulerWorkerCount)
	//return nil, nil
}
