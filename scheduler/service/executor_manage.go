package service

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/executor_operator"
	"sync"
	"time"
)

type ExecutorManageService struct {
	shutdownCh chan struct{}

	//key为executor的instanceID
	executors         map[string]*ExecutorWrapper
	statisticsService *StatisticsService
	discoveryClient   discovery.Client

	updateExecutorListeners []UpdateExecutorListener
}

func NewExecutorManageService(statisticsService *StatisticsService, discoveryClient discovery.Client) *ExecutorManageService {
	return &ExecutorManageService{
		shutdownCh:              make(chan struct{}),
		executors:               make(map[string]*ExecutorWrapper),
		statisticsService:       statisticsService,
		discoveryClient:         discoveryClient,
		updateExecutorListeners: make([]UpdateExecutorListener, 0),
	}
}

type UpdateExecutorListener interface {
	OnExecutorUpdate(newExecutors map[string]*ExecutorWrapper)
}

type ExecutorWrapper struct {
	Executor *model.Executor
	Operator executor_operator.Operator
	Status   *model.ExecutorStatus
}

func (s *ExecutorManageService) HeartBeat() {
	updateExecutorTicker := time.NewTicker(s.statisticsService.GetExecutorHeartbeatInterval())

	for {
		select {
		case <-s.shutdownCh:
			updateExecutorTicker.Stop()
			return
		case <-updateExecutorTicker.C:
			s.updateExecutor()
			updateExecutorTicker.Reset(s.statisticsService.GetExecutorHeartbeatInterval())
		}
	}
}

func (s *ExecutorManageService) Stop() {
	s.shutdownCh <- struct{}{}
}

func (s *ExecutorManageService) updateExecutor() {
	newServiceInstances := s.discoveryClient.DiscoverServices(constance.ExecutorServiceName)
	newExecutors := make(map[string]*ExecutorWrapper, len(newServiceInstances))

	for _, newInstance := range newServiceInstances {
		//合并新旧executor
		oldInstance, ok := s.executors[newInstance.InstanceId]

		if ok && oldInstance.Executor.Port == newInstance.Port && oldInstance.Executor.Host == newInstance.Host {
			//如果该executor上一波就有，那么继续用旧的
			newExecutors[newInstance.InstanceId] = oldInstance
			continue
		}

		//是新的executor，需要重新创建连接和Executor实例
		newExecutor, err := model.NewExecutorFromServiceInstance(newInstance)
		if err != nil {
			panic(err)
		}

		operator, err := executor_operator.NewOperatorByProtoc(newExecutor.Protoc, newExecutor.Host, newExecutor.Port)
		if err != nil {
			panic(err)
		}

		newExecutors[newInstance.InstanceId] = &ExecutorWrapper{
			Executor: newExecutor,
			Operator: operator,
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(newExecutors))

	type ret struct {
		instanceID string
		status     *model.ExecutorStatus
		err        error
	}

	rets := make([]*ret, len(newExecutors))
	counter := 0
	for _, newExecutor := range newExecutors {
		go func(e *ExecutorWrapper, retIdx int) {
			defer wg.Done()
			status, err := e.Operator.CheckStatus(s.statisticsService.GetHeartBeatTimeout())
			rets[retIdx] = &ret{
				instanceID: e.Executor.InstanceId,
				status:     status,
				err:        err,
			}
		}(newExecutor, counter)
		counter++
	}
	wg.Wait()

	for _, r := range rets {
		if r.err != nil {
			delete(newExecutors, r.instanceID)
			klog.Errorf("updateExecutor executor:%v failed with error:%v", r.instanceID, r.err)
		} else {
			klog.Infof("updateExecutor executor:%v success, status:%+v", r.instanceID, r.status)
		}
	}

	s.executors = newExecutors

	for _, l := range s.updateExecutorListeners {
		l.OnExecutorUpdate(s.executors)
	}
}

func (s *ExecutorManageService) AddUpdateExecutorListener(l UpdateExecutorListener) {
	s.updateExecutorListeners = append(s.updateExecutorListeners, l)
}
