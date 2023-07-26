// Code generated by Kitex v0.6.1. DO NOT EDIT.

package executor

import (
	"context"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	transport "github.com/cloudwego/kitex/transport"
	"supernova/pkg/api"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	HeartBeat(ctx context.Context, Req *api.HeartBeatRequest, callOptions ...callopt.Option) (r *api.HeartBeatResponse, err error)
	RunJob(ctx context.Context, callOptions ...callopt.Option) (stream Executor_RunJobClient, err error)
}

type Executor_RunJobClient interface {
	streaming.Stream
	Send(*api.RunJobRequest) error
	Recv() (*api.RunJobResponse, error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, client.WithTransportProtocol(transport.GRPC))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfo(), options...)
	if err != nil {
		return nil, err
	}
	return &kExecutorClient{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewClient creates a client for the service defined in IDL. It panics if any error occurs.
func MustNewClient(destService string, opts ...client.Option) Client {
	kc, err := NewClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kExecutorClient struct {
	*kClient
}

func (p *kExecutorClient) HeartBeat(ctx context.Context, Req *api.HeartBeatRequest, callOptions ...callopt.Option) (r *api.HeartBeatResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.HeartBeat(ctx, Req)
}

func (p *kExecutorClient) RunJob(ctx context.Context, callOptions ...callopt.Option) (stream Executor_RunJobClient, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.RunJob(ctx)
}
