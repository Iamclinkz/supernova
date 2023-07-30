package dal

import (
	"supernova/pkg/api/executor"
	"supernova/pkg/constance"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
)

type ExecutorRpcClient struct {
	cli executor.Client
}

func NewExecutorServiceClient(host, port string) (*ExecutorRpcClient, error) {
	if cli, err := executor.NewClient(
		constance.ExecutorServiceName,
		client.WithHostPorts(host+":"+port),
		//client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: constance.SchedulerServiceName}),
		//client.WithMiddleware(middleware.PrintKitexRequestResponse),
		client.WithTransportProtocol(transport.GRPC)); err != nil {
		return nil, err
	} else {
		return &ExecutorRpcClient{cli: cli}, nil
	}
}

func (c *ExecutorRpcClient) Cli() executor.Client {
	return c.cli
}
