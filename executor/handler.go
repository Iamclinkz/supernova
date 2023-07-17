package main

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/pkg/api"
)

func (e *Executor) HeartBeat(ctx context.Context, req *api.HeartBeatRequest) (resp *api.HeartBeatResponse, err error) {
	resp = new(api.HeartBeatResponse)
	resp.Cpu = 100
	resp.Memory = 100
	resp.RunningJob = 100
	return
}

func (e *Executor) RunJob(ctx context.Context, req *api.RunJobRequest) (resp *api.RunJobResponse, err error) {
	klog.CtxInfof(ctx, "fake run job: %+v", req)
	resp.Ok = true
	resp.Result = "fakeResult"
	return
}
