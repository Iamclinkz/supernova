package handler

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/executor/service"
	"supernova/pkg/api"
	"sync"
	"sync/atomic"
)

type GrpcHandler struct {
	executeService    *service.ExecuteService
	statisticsService *service.StatisticsService
}

func NewGrpcHandler(executeService *service.ExecuteService, statisticsService *service.StatisticsService) *GrpcHandler {
	return &GrpcHandler{
		executeService:    executeService,
		statisticsService: statisticsService,
	}
}

// HeartBeat implements the ExecutorImpl interface.
func (e *GrpcHandler) HeartBeat(ctx context.Context, req *api.HeartBeatRequest) (resp *api.HeartBeatResponse, err error) {
	resp = new(api.HeartBeatResponse)
	resp.HealthStatus = e.statisticsService.GetStatus()
	return
}

func (e *GrpcHandler) RunJob(stream api.Executor_RunJobServer) (err error) {
	//todo 优雅退出处理的有点粗糙，有时间再看看
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
				klog.Errorf("run job error:%v", connError)
				stop.Store(true)
				wg.Done()
				return
			}
			e.executeService.PushJobRequest(req)
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

			connError = stream.Send(resp)
			if connError != nil {
				//再把这个JobResult扔回去，让别的Scheduler发送
				e.executeService.PushJobResponse(resp)
				stop.Store(true)
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
	return nil
}
