package model

import (
	"errors"
	"fmt"
	"supernova/pkg/discovery"
	"supernova/scheduler/util"
)

type Executor struct {
	*discovery.ServiceInstance
	Tags []string
}

func NewExecutorFromServiceInstance(instance *discovery.ServiceInstance) (*Executor, error) {
	if instance == nil || instance.InstanceId == "" || instance.Meta == nil || instance.Host == "" ||
		instance.Port <= 0 || instance.Protoc == "" || instance.Meta[discovery.ExecutorTagFieldName] == "" {
		return nil, errors.New(fmt.Sprintf("can not decode executor service from:%+v", instance))
	}

	return &Executor{
		ServiceInstance: instance,
		Tags:            util.DecodeTags(instance.Meta[discovery.ExecutorTagFieldName]),
	}, nil
}
