package service

import (
	"supernova/pkg/discovery"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/executor_operator"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type ExecutorManageService struct {
	shutdownCh chan struct{}

	executorLock     sync.Mutex           //CheckExecutorAlive和updateExecutor，对executors的访问互斥的锁
	checkSequenceNum int                  //感觉环回也无所谓，暂时32位吧
	executors        map[string]*Executor //key为executor的instanceID

	statisticsService *StatisticsService
	discoveryClient   discovery.ExecutorDiscoveryClient

	updateExecutorListeners     []UpdateExecutorListener
	onJobResponseNotifyFuncFunc executor_operator.OnJobResponseNotifyFunc
}

func NewExecutorManageService(statisticsService *StatisticsService, discoveryClient discovery.ExecutorDiscoveryClient) *ExecutorManageService {
	return &ExecutorManageService{
		shutdownCh:              make(chan struct{}),
		executors:               make(map[string]*Executor),
		statisticsService:       statisticsService,
		discoveryClient:         discoveryClient,
		updateExecutorListeners: make([]UpdateExecutorListener, 0),
		executorLock:            sync.Mutex{},
	}
}

type UpdateExecutorListener interface {
	OnExecutorUpdate(newExecutors map[string]*Executor)
}

type Executor struct {
	ServiceData *discovery.ExecutorServiceInstance
	Operator    executor_operator.Operator
	Status      *model.ExecutorStatus
}

func (s *ExecutorManageService) HeartBeat() {
	updateExecutorTicker := time.NewTicker(1 * time.Millisecond)

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

// updateExecutor 根据从服务发现接口拿到的新的executor服务，跟executor建立连接，并且存储连接。
func (s *ExecutorManageService) updateExecutor() {
	//updateExecutor 本身就带有检查alive的能力，所以updateExecutor期间直接不让CheckExecutorAlive跑了
	s.executorLock.Lock()
	defer s.executorLock.Unlock()

	newServiceInstances := s.discoveryClient.DiscoverServices()
	newExecutors := make(map[string]*Executor, len(newServiceInstances))

	// 这里注意，因为优雅退出，所以假设老的ExecutorA的连接仍然可用，但是新从服务发现接口中拿到的实例中没有ExecutorA，
	// 则认为ExecutorA准备优雅退出了。这时候Scheduler就不给ExecutorA发消息，让ExecutorA处理任务了
	//（具体实现是在ExecutorManagerService中缓存的executors中没有老的ExecutorA了）
	// 但是Scheduler <-> ExecutorA的连接仍然不会断开
	for _, newInstanceServiceData := range newServiceInstances {
		//合并新旧executor
		if newInstanceServiceData == nil {
			panic("")
		}
		oldInstance, ok := s.executors[newInstanceServiceData.InstanceId]

		if ok && oldInstance.ServiceData.Port == newInstanceServiceData.Port &&
			oldInstance.ServiceData.Host == newInstanceServiceData.Host && oldInstance.Operator.Alive() {
			//如果该executor上一波就有，且连接还能用，那么继续用旧的
			newExecutors[newInstanceServiceData.InstanceId] = oldInstance
			continue
		}

		operator, err := executor_operator.NewOperatorByProtoc(newInstanceServiceData.Protoc,
			newInstanceServiceData.Host, newInstanceServiceData.Port, s.onJobResponseNotifyFuncFunc)
		if err != nil {
			panic(err)
		}

		newExecutors[newInstanceServiceData.InstanceId] = &Executor{
			ServiceData: newInstanceServiceData,
			Operator:    operator,
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
		go func(e *Executor, retIdx int) {
			defer wg.Done()
			status, err := e.Operator.CheckStatus(s.statisticsService.GetHeartBeatTimeout())
			rets[retIdx] = &ret{
				instanceID: e.ServiceData.InstanceId,
				status:     status,
				err:        err,
			}
		}(newExecutor, counter)
		counter++
	}
	wg.Wait()

	for _, r := range rets {
		if r.err != nil || r.status.GracefulStopped {
			//如果Executor已经优雅退出了，那么删掉
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
// 经过常哥的指点，Scheduler和Executor的连接如果单方面断开，就可以认为连接不可用了。直接退出连接处理即可，不需要再去检查。
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
				klog.Errorf("find executorID:%v dead", instanceID)
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
