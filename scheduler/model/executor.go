package model

import (
	"errors"
	"fmt"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/pkg/util"
)

type Executor struct {
	*discovery.ServiceInstance
	Tags []string
}

func NewExecutorFromServiceInstance(instance *discovery.ServiceInstance) (*Executor, error) {
	if instance == nil || instance.InstanceId == "" || instance.Meta == nil || instance.Host == "" ||
		instance.Port <= 0 || instance.Protoc == "" {
		return nil, errors.New(fmt.Sprintf("can not decode executor service from:%+v", instance))
	}

	return &Executor{
		ServiceInstance: instance,
		Tags:            util.DecodeTags(instance.Meta[constance.ExecutorTagFieldName]),
	}, nil
}
