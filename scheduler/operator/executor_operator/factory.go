package executor_operator

import (
	"errors"
	"supernova/pkg/api"
	"supernova/pkg/discovery"
	"supernova/scheduler/model"
	"time"
)

type OnJobResponseNotifyFunc func(response *api.RunJobResponse)

type Operator interface {
	CheckStatus(timeout time.Duration) (*model.ExecutorStatus, error)
	RunJob(request *api.RunJobRequest) (err error)
	Alive() bool
}

func NewOperatorByProtoc(protoc discovery.ProtocType, host string, port int, f OnJobResponseNotifyFunc) (Operator, error) {
	switch protoc {
	case discovery.ProtocTypeGrpc:
		return newGrpcOperator(host, port, f)
	case discovery.ProtocTypeHttp:
		return newHttpOperator(host, port, f)
	default:
		return nil, errors.New("missing protoc")
	}
}
