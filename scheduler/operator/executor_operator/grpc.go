package executor_operator

import (
	"context"
	"errors"
	"strconv"
	"supernova/pkg/api"
	"supernova/pkg/api/executor"
	"supernova/scheduler/dal"
	"supernova/scheduler/model"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

type GrpcOperator struct {
	executorCli *dal.ExecutorRpcClient
	notify      OnJobResponseNotifyFunc
	stream      executor.Executor_RunJobClient
	streamLock  sync.Mutex
	stop        atomic.Bool

	//for debug
	host string
	port int
}

func (g *GrpcOperator) Alive() bool {
	return !g.stop.Load()
}

func newGrpcOperator(host string, port int, f OnJobResponseNotifyFunc) (Operator, error) {
	if client, err := dal.NewExecutorServiceClient(host, strconv.Itoa(port)); err != nil {
		return nil, err
	} else {
		return &GrpcOperator{
			executorCli: client,
			streamLock:  sync.Mutex{},
			host:        host,
			port:        port,
			notify:      f,
		}, nil
	}
}

func (g *GrpcOperator) CheckStatus(timeout time.Duration) (*model.ExecutorStatus, error) {
	req := new(api.HeartBeatRequest)
	resp, err := g.executorCli.Cli().HeartBeat(context.TODO(), req, callopt.WithRPCTimeout(timeout))

	if err != nil {
		g.ForceStop()
		return nil, err
	}

	status := new(model.ExecutorStatus)
	status.FromPb(resp)
	return status, nil
}

func (g *GrpcOperator) RunJob(request *api.RunJobRequest) error {
	if !g.Alive() {
		//如果流已经停止了，那么直接返回错误
		return errors.New("stream stopped")
	}

	if g.stream == nil {
		//如果流是空的，那么创建流
		if err := g.initStream(); err != nil {
			return err
		}
	}

	var err error
	if err = g.stream.Send(request); err != nil {
		g.ForceStop()
	}
	return err
}

func (g *GrpcOperator) initStream() error {
	var err error
	//加锁避免重复获取流
	g.streamLock.Lock()
	defer g.streamLock.Unlock()

	if g.stream != nil {
		return nil
	}

	if g.stream, err = g.executorCli.Cli().RunJob(context.TODO()); err != nil {
		return err
	}

	//todo 删掉判断
	if g.notify == nil {
		panic("")
	}

	//开一个go程专门接收stream的消息
	go func() {
		var (
			receiveErr     error
			runJobResponse *api.RunJobResponse
		)
		klog.Infof("init a stream between %v:%v", g.host, g.port)
		for g.Alive() {
			runJobResponse, receiveErr = g.stream.Recv()
			if receiveErr != nil {
				g.ForceStop()
				return
			}

			g.notify(runJobResponse)
		}
	}()

	return nil
}

func (g *GrpcOperator) RegisterJobResponseNotify(f OnJobResponseNotifyFunc) {
	g.notify = f
}

func (g *GrpcOperator) ForceStop() {
	g.stop.Store(true)
}

var _ Operator = (*GrpcOperator)(nil)
