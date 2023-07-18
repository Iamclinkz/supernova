package service

import (
	"errors"
	"math/rand"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/util"
	"sync"
)

type ExecutorRouteService struct {
	executorManagerService *ExecutorManageService
	jobRouters             *JobRouterManager

	mu sync.Mutex
	//key为tag，value为有该tag的Executor
	invertedIndex map[string][]*ExecutorWrapper
	executors     map[string]*ExecutorWrapper
}

func NewExecutorRouteService(executorManagerService *ExecutorManageService) *ExecutorRouteService {
	ret := &ExecutorRouteService{
		executorManagerService: executorManagerService,
		jobRouters:             NewJobRouters(),
		mu:                     sync.Mutex{},
		invertedIndex:          make(map[string][]*ExecutorWrapper),
		executors:              make(map[string]*ExecutorWrapper),
	}

	executorManagerService.AddUpdateExecutorListener(ret)
	return ret
}

func (s *ExecutorRouteService) ChooseJobExecutor(job *model.Job) (*ExecutorWrapper, error) {
	executors, err := s.findMatchingExecutorsForJob(job)
	if err != nil {
		return nil, err
	}

	return s.jobRouters.Route(job.ExecutorRouteStrategy, executors, job)
}

// findMatchingExecutorsForJob 找到和Job的Tag匹配的Executor
func (s *ExecutorRouteService) findMatchingExecutorsForJob(job *model.Job) ([]*ExecutorWrapper, error) {
	ret := make([]*ExecutorWrapper, 0)
	tags := util.DecodeTags(job.Tags)
	if len(tags) == 0 {
		return nil, errors.New("no tag in job")
	}

	executorCount := make(map[string]int, len(tags))

	s.mu.Lock()
	executors := s.executors
	invertedIndex := s.invertedIndex
	s.mu.Unlock()

	for _, tag := range tags {
		for _, executor := range invertedIndex[tag] {
			executorCount[executor.Executor.InstanceId]++
		}
	}

	for instanceID, count := range executorCount {
		if count == len(tags) {
			ret = append(ret, executors[instanceID])
		}
	}

	if len(ret) == 0 {
		return nil, errors.New("can not find exist executor for job")
	}

	return ret, nil
}

// OnExecutorUpdate 在更新Executor时，刷新当前的倒排索引
func (s *ExecutorRouteService) OnExecutorUpdate(newExecutors map[string]*ExecutorWrapper) {
	//todo 注册到executorManageService中
	newInvertedIndex := make(map[string][]*ExecutorWrapper, len(newExecutors))
	for _, newExecutor := range newExecutors {
		for _, tag := range newExecutor.Executor.Tags {
			newInvertedIndex[tag] = append(newInvertedIndex[tag], newExecutor)
		}
	}

	s.mu.Lock()
	s.invertedIndex = newInvertedIndex
	s.executors = newExecutors
	s.mu.Unlock()
}

type Router interface {
	Route(instance []*ExecutorWrapper, job *model.Job) *ExecutorWrapper
}

type JobRouterManager struct {
	routers map[constance.ExecutorRouteStrategyType]Router
}

func NewJobRouters() *JobRouterManager {
	return &JobRouterManager{routers: map[constance.ExecutorRouteStrategyType]Router{
		constance.ExecutorRouteStrategyTypeRandom: RandomRouter{},
	}}
}

func (r *JobRouterManager) Route(strategy constance.ExecutorRouteStrategyType, instance []*ExecutorWrapper, job *model.Job) (*ExecutorWrapper, error) {
	if router, ok := r.routers[strategy]; ok {
		return router.Route(instance, job), nil
	}

	return nil, errors.New("no ExecutorRouteStrategyType:" + strategy.String())
}

// RandomRouter 随机路由
type RandomRouter struct{}

func (r RandomRouter) Route(instance []*ExecutorWrapper, job *model.Job) *ExecutorWrapper {
	if len(instance) == 0 {
		return nil
	}

	return instance[rand.Int()%len(instance)]
}

var _ Router = (*RandomRouter)(nil)
