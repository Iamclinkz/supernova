package service

import (
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/executor_operator"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type ExecutorManageService struct {
	shutdownCh chan struct{}

	executorLock     sync.Mutex                  //CheckExecutorAlive和updateExecutor，对executors的访问互斥的锁
	checkSequenceNum int                         //感觉环回也无所谓，暂时32位吧
	executors        map[string]*ExecutorWrapper //key为executor的instanceID

	statisticsService *StatisticsService
	discoveryClient   discovery.Client

	updateExecutorListeners     []UpdateExecutorListener
	onJobResponseNotifyFuncFunc executor_operator.OnJobResponseNotifyFunc
}

func NewExecutorManageService(statisticsService *StatisticsService, discoveryClient discovery.Client) *ExecutorManageService {
	return &ExecutorManageService{
		shutdownCh:              make(chan struct{}),
		executors:               make(map[string]*ExecutorWrapper),
		statisticsService:       statisticsService,
		discoveryClient:         discoveryClient,
		updateExecutorListeners: make([]UpdateExecutorListener, 0),
		executorLock:            sync.Mutex{},
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
	//updateExecutor 本身就带有检查alive的能力，所以updateExecutor期间直接不让CheckExecutorAlive跑了
	s.executorLock.Lock()
	defer s.executorLock.Unlock()

	newServiceInstances := s.discoveryClient.DiscoverServices(constance.ExecutorServiceName)
	newExecutors := make(map[string]*ExecutorWrapper, len(newServiceInstances))

	for _, newInstance := range newServiceInstances {
		//合并新旧executor
		if newInstance == nil {
			panic("")
		}
		oldInstance, ok := s.executors[newInstance.InstanceId]

		if ok && oldInstance.Executor.Port == newInstance.Port &&
			oldInstance.Executor.Host == newInstance.Host && oldInstance.Operator.Alive() {
			//如果该executor上一波就有，且连接还能用，那么继续用旧的
			newExecutors[newInstance.InstanceId] = oldInstance
			continue
		}

		//是新的executor，需要重新创建连接和Executor实例
		newExecutor, err := model.NewExecutorFromServiceInstance(newInstance)
		if err != nil {
			panic(err)
		}

		operator, err := executor_operator.NewOperatorByProtoc(newExecutor.Protoc,
			newExecutor.Host, newExecutor.Port, s.onJobResponseNotifyFuncFunc)
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

	s.NotifyExecutorListener()
	s.checkSequenceNum++
}

// CheckExecutorAlive 其他service发现某个executor有问题，让manager看看要不要删掉
func (s *ExecutorManageService) CheckExecutorAlive(instanceID string) {
	s.executorLock.Lock()
	unhealthyExecutor := s.executors[instanceID]
	seqNum := s.checkSequenceNum
	if unhealthyExecutor == nil {
		//如果根本没有instance，不用Check
		s.executorLock.Unlock()
		return
	}
	if !unhealthyExecutor.Operator.Alive() {
		//如果确实狗带了，那么删除掉
		delete(s.executors, instanceID)
	}
	s.executorLock.Unlock()

	//代码执行到这里，其他地方反应狗带了，但是还是Alive的，那么需要检查
	go func() {
		_, err := unhealthyExecutor.Operator.CheckStatus(s.statisticsService.GetHeartBeatTimeout())
		if err != nil {
			//如果确实不健康
			s.executorLock.Lock()
			if s.checkSequenceNum == seqNum {
				//如果已经updateExecutor更新过一次了，那么自己不更新了
				delete(s.executors, instanceID)
				//虽然在listener也做了幂等，但是还是加锁吧
				s.NotifyExecutorListener()
			}
			s.executorLock.Unlock()
		}
	}()
}

func (s *ExecutorManageService) AddUpdateExecutorListener(l UpdateExecutorListener) {
	s.updateExecutorListeners = append(s.updateExecutorListeners, l)
}

func (s *ExecutorManageService) NotifyExecutorListener() {
	for _, l := range s.updateExecutorListeners {
		l.OnExecutorUpdate(s.executors)
	}
}

func (s *ExecutorManageService) RegisterReceiveMsgNotifyFunc(f executor_operator.OnJobResponseNotifyFunc) {
	if s.onJobResponseNotifyFuncFunc != nil {
		//todo 测试使用
		panic("")
	}

	s.onJobResponseNotifyFuncFunc = f
}
