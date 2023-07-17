package executor_operator

import (
	"context"
	"github.com/cloudwego/kitex/client/callopt"
	"strconv"
	"supernova/pkg/api"
	"supernova/scheduler/dal"
	"supernova/scheduler/model"
	"time"
)

type GrpcOperator struct {
	cli *dal.ExecutorRpcClient
}

func newGrpcOperator(host string, port int) (Operator, error) {
	if client, err := dal.NewExecutorServiceClient(host, strconv.Itoa(port)); err != nil {
		return nil, err
	} else {
		return &GrpcOperator{cli: client}, nil
	}
}

func (g *GrpcOperator) CheckStatus(timeout time.Duration) (*model.ExecutorStatus, error) {
	req := new(api.HeartBeatRequest)
	resp, err := g.cli.Cli().HeartBeat(context.TODO(), req, callopt.WithRPCTimeout(timeout))

	if err != nil {
		return nil, err
	}

	status := new(model.ExecutorStatus)
	status.FromPb(resp)
	return status, nil
}

func (g *GrpcOperator) RunJob(request *model.RunJobRequest, timeout time.Duration) (*model.RunJobResponse, error) {
	pbReq := request.ToPb()
	pbResp, err := g.cli.Cli().RunJob(context.TODO(), pbReq, callopt.WithRPCTimeout(timeout))

	if err != nil {
		return nil, err
	}

	resp := new(model.RunJobResponse)
	resp.FromPb(pbResp)
	return resp, nil
}

var _ Operator = (*GrpcOperator)(nil)
