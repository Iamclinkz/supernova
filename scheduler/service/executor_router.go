package service

import (
	"errors"
	"fmt"
	"math/rand"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"sync"
)

type ExecutorRouteService struct {
	executorManagerService *ExecutorManageService
	jobRouterManager       *JobRouterManager

	mu sync.Mutex
	//key为tag，value为有该tag的Executor
	invertedIndex map[string][]*ExecutorWrapper
	executors     map[string]*ExecutorWrapper
}

func NewExecutorRouteService(executorManagerService *ExecutorManageService) *ExecutorRouteService {
	ret := &ExecutorRouteService{
		executorManagerService: executorManagerService,
		jobRouterManager:       NewJobRouters(),
		mu:                     sync.Mutex{},
		invertedIndex:          make(map[string][]*ExecutorWrapper),
		executors:              make(map[string]*ExecutorWrapper),
	}

	executorManagerService.AddUpdateExecutorListener(ret)
	return ret
}

func (s *ExecutorRouteService) ChooseJobExecutor(job *model.Job, onFireLog *model.OnFireLog, retry bool) (*ExecutorWrapper, error) {
	executors, err := s.findMatchingExecutorsForJob(job)
	if err != nil {
		return nil, err
	}

	if retry && onFireLog.ExecutorInstance != "" {
		//重试的话，暂时只能分配到同一个executor节点。否则失败
		//todo： 以后要不要加上“至少执行一次”语义，如果分配不到上次执行的Executor，那么分配新的Executor执行？
		return s.jobRouterManager.Route(constance.ExecutorRouteStrategyTypeExecutorInstanceMatch, executors, job, onFireLog)
	}
	return s.jobRouterManager.Route(job.ExecutorRouteStrategy, executors, job, onFireLog)
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
		return nil, fmt.Errorf("can not find exist executor for job, executors:%v", executors)
	}

	return ret, nil
}

// OnExecutorUpdate 在更新Executor时，刷新当前的倒排索引
func (s *ExecutorRouteService) OnExecutorUpdate(newExecutors map[string]*ExecutorWrapper) {
	newInvertedIndex := make(map[string][]*ExecutorWrapper, len(newExecutors))
	for _, newExecutor := range newExecutors {
		for _, tag := range newExecutor.Executor.Tags {
			newInvertedIndex[tag] = append(newInvertedIndex[tag], newExecutor)
		}
	}

	//加锁保证这俩一起更新
	s.mu.Lock()
	s.invertedIndex = newInvertedIndex
	s.executors = newExecutors
	s.mu.Unlock()
}

type Router interface {
	Route(instances []*ExecutorWrapper, job *model.Job, onFireLog *model.OnFireLog) *ExecutorWrapper
}

type JobRouterManager struct {
	routers map[constance.ExecutorRouteStrategyType]Router
}

func NewJobRouters() *JobRouterManager {
	return &JobRouterManager{routers: map[constance.ExecutorRouteStrategyType]Router{
		constance.ExecutorRouteStrategyTypeRandom:                RandomRouter{},
		constance.ExecutorRouteStrategyTypeExecutorInstanceMatch: ExecutorInstanceIDRouter{},
	}}
}

func (r *JobRouterManager) Route(strategy constance.ExecutorRouteStrategyType,
	instance []*ExecutorWrapper, job *model.Job, onFireLog *model.OnFireLog) (*ExecutorWrapper, error) {
	if router, ok := r.routers[strategy]; ok {
		return router.Route(instance, job, onFireLog), nil
	}

	return nil, errors.New("no ExecutorRouteStrategyType:" + strategy.String())
}

// RandomRouter 随机路由
type RandomRouter struct{}

func (r RandomRouter) Route(instances []*ExecutorWrapper, job *model.Job, onFireLog *model.OnFireLog) *ExecutorWrapper {
	if len(instances) == 0 {
		return nil
	}

	return instances[rand.Int()%len(instances)]
}

// ExecutorInstanceIDRouter 按照Executor的InstanceID路由
type ExecutorInstanceIDRouter struct{}

func (r ExecutorInstanceIDRouter) Route(instances []*ExecutorWrapper, job *model.Job, onFireLog *model.OnFireLog) *ExecutorWrapper {
	if len(instances) == 0 {
		return nil
	}

	for _, instance := range instances {
		if instance.Executor.InstanceId == onFireLog.ExecutorInstance {
			return instance
		}
	}
	return instances[rand.Int()%len(instances)]
}

var _ Router = (*RandomRouter)(nil)
