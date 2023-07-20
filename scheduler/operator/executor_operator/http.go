package executor_operator

import (
	"strconv"
	"supernova/pkg/api"
	"supernova/scheduler/model"
	"time"
)

type HttpOperator struct {
	host string
	port string
}

func (h HttpOperator) Alive() bool {
	//TODO implement me
	panic("implement me")
}

func newHttpOperator(host string, port int) (Operator, error) {
	return &HttpOperator{
		host: host,
		port: strconv.Itoa(port),
	}, nil
}

func (h HttpOperator) CheckStatus(timeout time.Duration) (*model.ExecutorStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (h HttpOperator) RunJob(*api.RunJobRequest) error {
	//TODO implement me
	panic("implement me")
	return nil
}

var _ Operator = (*HttpOperator)(nil)
