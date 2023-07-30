package util

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type MapCarrier map[string]string

func (mc MapCarrier) Get(key string) string {
	return mc[key]
}

func (mc MapCarrier) Set(key, value string) {
	mc[key] = value
}

func (mc MapCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range mc {
		keys = append(keys, k)
	}
	return keys
}

func GenTraceContext(ctx context.Context) map[string]string {
	var ret MapCarrier = make(map[string]string, 0)
	otel.GetTextMapPropagator().Inject(ctx, ret)
	return ret
}

func TraceCtx2String(ctx context.Context) string {
	var tmp MapCarrier = make(map[string]string, 0)
	otel.GetTextMapPropagator().Inject(ctx, tmp)
	b, err := json.Marshal(tmp)
	if err != nil {
		//todo 删掉
		panic(err)
	}
	return string(b)
}

func String2TraceCtx(traceContext string) context.Context {
	var tmp MapCarrier
	err := json.Unmarshal([]byte(traceContext), &tmp)
	if err != nil {
		//todo 删掉
		panic(err)
	}

	return otel.GetTextMapPropagator().Extract(context.Background(), tmp)
}

func NewSpanFromTraceContext(name string, tracer trace.Tracer, traceContext map[string]string) (context.Context, trace.Span) {
	traceCtx := otel.GetTextMapPropagator().Extract(context.Background(), MapCarrier(traceContext))
	return tracer.Start(traceCtx, name)
}
