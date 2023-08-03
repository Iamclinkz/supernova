package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"supernova/pkg/conf"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/pkg/session/trace"
	"supernova/scheduler/dal"
	"supernova/scheduler/operator/schedule_operator"
	"supernova/scheduler/operator/schedule_operator/memory_operator"
	"supernova/scheduler/operator/schedule_operator/mysql_operator"
	"supernova/scheduler/operator/schedule_operator/redis_operator"
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/google/uuid"
)

type SchedulerBuilder struct {
	instanceID           string
	scheduleOperator     schedule_operator.Operator
	discoveryClient      discovery.DiscoverClient
	schedulerWorkerCount int
	tracerProvider       *sdktrace.TracerProvider
	meterProvider        *metric.MeterProvider
	err                  error
	standalone           bool
	oTelConfig           *trace.OTelConfig
}

func NewSchedulerBuilder() *SchedulerBuilder {
	return &SchedulerBuilder{}
}

func (b *SchedulerBuilder) WithConsulDiscovery(consulHost, consulPort string) *SchedulerBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(
		discovery.TypeConsul,
		discovery.NewConsulMiddlewareConfig(consulHost, consulPort),
		discovery.NewConsulRegisterConfig("", true))
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
	b.scheduleOperator = memory_operator.NewMemoryScheduleOperator()
	return b
}

func (b *SchedulerBuilder) WithRedisStore() *SchedulerBuilder {
	cli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "password",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	_, err := cli.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	b.scheduleOperator = redis_operator.NewRedisScheduleOperator(cli)
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

func (b *SchedulerBuilder) WithOTelConfig(oTelConfig *trace.OTelConfig) *SchedulerBuilder {
	if oTelConfig.EnableTrace || oTelConfig.EnableMetrics {
		var err error
		b.tracerProvider, b.meterProvider, err = trace.InitProvider(constance.SchedulerServiceName, oTelConfig.InstrumentConf)
		if err != nil && b.err != nil {
			b.err = err
		}
	}

	//来不及改了。。先这样吧
	if !oTelConfig.EnableTrace {
		b.tracerProvider = nil
	}
	if !oTelConfig.EnableMetrics {
		b.meterProvider = nil
	}
	b.oTelConfig = oTelConfig
	return b
}

func (b *SchedulerBuilder) WithStandalone() *SchedulerBuilder {
	b.standalone = true
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
		b.schedulerWorkerCount = 128
	}
	if b.instanceID == "" {
		b.instanceID = fmt.Sprintf("Scheduler-%v", uuid.New())
	}

	return genScheduler(b.instanceID, b.oTelConfig,
		b.tracerProvider, b.meterProvider, b.scheduleOperator, b.discoveryClient, b.schedulerWorkerCount, b.standalone)
}
