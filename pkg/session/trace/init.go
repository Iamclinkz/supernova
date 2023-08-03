package trace

import (
	"context"
	"fmt"
	"supernova/pkg/conf"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type OTelConfig struct {
	EnableTrace    bool
	EnableMetrics  bool
	InstrumentConf *conf.OTelConf
}

func InitProvider(serviceName string, instrumentConf *conf.OTelConf) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx,
		fmt.Sprintf("%v:%v", instrumentConf.CollectorEndpointHost, instrumentConf.CollectorEndpointPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		//todo 改一下
		sdktrace.WithSampler(sdktrace.NeverSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricsExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	readerOption := sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricsExporter, sdkmetric.WithInterval(5*time.Second)))

	meterProvider := sdkmetric.NewMeterProvider(
		readerOption,
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return tracerProvider, meterProvider, nil
}
