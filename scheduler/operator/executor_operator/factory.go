package executor_operator

import (
	"errors"
	"supernova/pkg/api"
	"supernova/pkg/discovery"
	"supernova/scheduler/model"
	"time"
)

type Operator interface {
	CheckStatus(timeout time.Duration) (*model.ExecutorStatus, error)
	RunJob(request *api.RunJobRequest) (err error)
	Alive() bool
}

func NewOperatorByProtoc(protoc discovery.ProtocType, host string, port int) (Operator, error) {
	switch protoc {
	case discovery.ProtocTypeGrpc:
		return newGrpcOperator(host, port)
	case discovery.ProtocTypeHttp:
		return newHttpOperator(host, port)
	default:
		return nil, errors.New("missing protoc")
	}
}
