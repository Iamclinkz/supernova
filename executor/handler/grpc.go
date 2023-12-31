package handler

import (
	"context"
	"errors"
	"supernova/executor/service"
	"supernova/pkg/api"
	"supernova/pkg/util"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/cloudwego/kitex/pkg/klog"
)

type GrpcHandler struct {
	executeService    *service.ExecuteService
	statisticsService *service.StatisticsService
	gracefulStopped   atomic.Bool
	enableOTel        bool
	tracer            trace.Tracer
	instanceID        string
}

func NewGrpcHandler(executeService *service.ExecuteService,
	statisticsService *service.StatisticsService, enableOTel bool, instanceID string) *GrpcHandler {
	ret := &GrpcHandler{
		executeService:    executeService,
		statisticsService: statisticsService,
		enableOTel:        enableOTel,
		instanceID:        instanceID,
	}

	if enableOTel {
		ret.tracer = otel.Tracer("GrpcTracer")
	}

	return ret
}

// HeartBeat implements the ExecutorImpl interface.
func (e *GrpcHandler) HeartBeat(ctx context.Context, req *api.HeartBeatRequest) (resp *api.HeartBeatResponse, err error) {
	resp = new(api.HeartBeatResponse)
	resp.HealthStatus = e.statisticsService.GetStatus()
	return
}

func (e *GrpcHandler) RunJob(stream api.Executor_RunJobServer) (err error) {
	if e.gracefulStopped.Load() {
		return errors.New("executor graceful stopped")
	}

	var stop atomic.Bool
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var (
			req       *api.RunJobRequest
			connError error
		)
		for !stop.Load() {
			req, connError = stream.Recv()
			if connError != nil {
				klog.Errorf("[%v] receive job request error:%v", e.instanceID, connError)
				stop.Store(true)
				wg.Done()
				return
			}
			e.statisticsService.OnReceiveRunJobRequest(req)

			doTrace := len(req.TraceContext) != 0 && e.enableOTel
			var pushJobRequestChSpan trace.Span
			if doTrace {
				_, pushJobRequestChSpan = util.NewSpanFromTraceContext("pushJobToRequestCh", e.tracer, req.TraceContext)
			}
			e.executeService.PushJobRequest(req)
			if doTrace {
				pushJobRequestChSpan.End()
			}
		}
	}()

	go func() {
		var (
			resp      *api.RunJobResponse
			connError error
			chOk      bool
		)
		for !stop.Load() {
			resp, chOk = e.executeService.PopJobResponse()
			if !chOk {
				//主动关掉了处理任务返回的chan，说明不处理了，应该主动关掉
				_ = stream.Close()
				stop.Store(true)
				return
			}

			doTrace := len(resp.TraceContext) != 0 && e.enableOTel
			var streamSendSpan trace.Span
			if doTrace {
				_, streamSendSpan = util.NewSpanFromTraceContext("sendToGrpcStream", e.tracer, resp.TraceContext)
			}

			connError = stream.Send(resp)
			if connError != nil {
				if doTrace {
					streamSendSpan.RecordError(connError)
					streamSendSpan.End()
				}
				klog.Warnf("[%v] send responase back failed with error:%v", e.instanceID, err)
				//再把这个JobResult扔回去，让别的Scheduler发送
				e.executeService.PushJobResponse(resp)
				stop.Store(true)
				wg.Done()
				return
			}

			if doTrace {
				streamSendSpan.End()
			}
			e.statisticsService.OnSendRunJobResponseSuccess(resp)
		}
	}()

	wg.Wait()
	return nil
}

func (e *GrpcHandler) OnGracefulStop() {
	e.gracefulStopped.Store(true)
}
