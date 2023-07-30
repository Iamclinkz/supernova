package app

import (
	"errors"
	"strconv"
	"supernova/executor/processor"
	"supernova/pkg/conf"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	tconf "supernova/pkg/session/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ExecutorBuilder struct {
	instanceID      string
	tags            []string
	processor       map[string]processor.JobProcessor
	serveConf       *discovery.ExecutorServiceServeConf
	processorCount  int
	discoveryClient discovery.ExecutorDiscoveryClient
	extraConf       map[string]string
	traceProvider   *sdktrace.TracerProvider
	err             error
}

func NewExecutorBuilder() *ExecutorBuilder {
	return &ExecutorBuilder{
		tags:      make([]string, 0),
		processor: make(map[string]processor.JobProcessor, 0),
		extraConf: make(map[string]string),
	}
}

func (b *ExecutorBuilder) WithInstanceID(id string) *ExecutorBuilder {
	if id == "" && b.err != nil {
		b.err = errors.New("empty instanceID")
	} else {
		b.instanceID = id
	}

	return b
}

// WithEnvTag 增加环境相关Tag
func (b *ExecutorBuilder) WithEnvTag(tag string) *ExecutorBuilder {
	if tag == "" && b.err == nil {
		b.err = errors.New("empty tag")
	} else {
		b.tags = append(b.tags, discovery.EnvTagPrefix+tag)
	}

	return b
}

// WithResourceTag 增加计算资源相关Tag
func (b *ExecutorBuilder) WithResourceTag(tag string) *ExecutorBuilder {
	if tag == "" && b.err == nil {
		b.err = errors.New("empty tag")
	} else {
		b.tags = append(b.tags, discovery.ResourceTagPrefix+tag)
	}

	return b
}

// WithCustomTag 增加用户自定义Tag
func (b *ExecutorBuilder) WithCustomTag(tag string) *ExecutorBuilder {
	if tag == "" && b.err == nil {
		b.err = errors.New("empty tag")
	} else {
		b.tags = append(b.tags, discovery.CustomTagPrefix+tag)
	}

	return b
}

func (b *ExecutorBuilder) WithConsulDiscovery(consulHost, consulPort string,
	healthCheckPort int) *ExecutorBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(
		discovery.TypeConsul,
		discovery.NewConsulMiddlewareConfig(consulHost, consulPort),
		discovery.NewConsulRegisterConfig(strconv.Itoa(healthCheckPort)),
	)
	if err != nil && b.err == nil {
		b.err = err
	} else {
		b.discoveryClient = discoveryClient
	}

	return b
}

func (b *ExecutorBuilder) WithK8sDiscovery(namespace, k8sCheckHealthPort string) *ExecutorBuilder {
	discoveryClient, err := discovery.NewDiscoveryClient(
		discovery.TypeK8s,
		discovery.NewK8sMiddlewareConfig(namespace),
		discovery.NewK8sRegisterConfig(k8sCheckHealthPort),
	)
	if err != nil && b.err == nil {
		b.err = err
	} else {
		b.discoveryClient = discoveryClient
	}

	return b
}

func (b *ExecutorBuilder) WithProcessor(p processor.JobProcessor) *ExecutorBuilder {
	glueType := p.GetGlueType()

	if glueType == "" && b.err == nil {
		b.err = errors.New("processor.GetGlueType() return nothing")
	}

	b.processor[glueType] = p
	b.tags = append(b.tags, discovery.GlueTypeTagPrefix+glueType)
	return b
}

func (b *ExecutorBuilder) WithGrpcServe(host string, port int) *ExecutorBuilder {
	b.serveConf = new(discovery.ExecutorServiceServeConf)
	b.serveConf.Host = host
	b.serveConf.Port = port
	b.serveConf.Protoc = discovery.ProtocTypeGrpc

	return b
}

func (b *ExecutorBuilder) WithProcessorCount(count int) *ExecutorBuilder {
	b.processorCount = count
	return b
}

func (b *ExecutorBuilder) WithOTelCollector(instrumentConf *conf.OTelConf) *ExecutorBuilder {
	var err error
	b.traceProvider, err = tconf.InitProvider(constance.ExecutorServiceName, instrumentConf)
	if err != nil && b.err != nil {
		b.err = err
	}
	return b
}

func (b *ExecutorBuilder) Build() (*Executor, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.instanceID == "" {
		return nil, errors.New("no instanceID")
	}

	if len(b.processor) == 0 {
		return nil, errors.New("no processor")
	}

	if b.serveConf == nil {
		return nil, errors.New("no serve conf")
	}

	if len(b.tags) == 0 {
		return nil, errors.New("no tags")
	}

	if b.discoveryClient == nil {
		return nil, errors.New("no selected service discovery")
	}

	return genExecutor(b.instanceID, b.traceProvider != nil, b.traceProvider, b.tags,
		b.processor, b.serveConf, b.processorCount, b.discoveryClient, b.extraConf)
}
