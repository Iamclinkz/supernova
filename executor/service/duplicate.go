package service

import (
	"strconv"
	"supernova/pkg/api"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// DuplicateService 当前主要有以下功能：
// 1. 防止执行成功的任务（同一个OnFireID）再次执行。
// 2. 防止同一个任务（同一个OnFireID）同时执行
// 一个任务从收到grpc/http请求，一直到处理结束的调用链应该是：
// OnReceiveRunJobRequest -> CheckDuplicateExecuteSuccessJob -> CheckDuplicateExecute -> OnStartExecute -> OnFinishExecute
type DuplicateService struct {
	c               *cache.Cache
	mu              sync.Mutex
	onFireID2Waiter map[uint][]chan bool
}

func NewDuplicateService() *DuplicateService {
	return &DuplicateService{
		c:               cache.New(5*time.Minute, 10*time.Minute),
		mu:              sync.Mutex{},
		onFireID2Waiter: map[uint][]chan bool{},
	}
}

func (s *DuplicateService) OnReceiveRunJobRequest(request *api.RunJobRequest) {

}

func (s *DuplicateService) OnStartExecute(task *Task) {

}

func (s *DuplicateService) OnFinishExecute(task *Task) {
	onFireID := uint(task.JobResponse.OnFireLogID)
	s.mu.Lock()
	defer s.mu.Unlock()
	waitChanSlice, ok := s.onFireID2Waiter[onFireID]
	if !ok {
		//todo 测试使用
		panic("")
	}

	if task.JobResponse.Result.Ok {
		//如果执行成功，给所有的等待的CheckDuplicateExecute返回false
		for _, waitCh := range waitChanSlice {
			waitCh <- false
		}
		//如果成功，还需要记录一下成功
		s.c.Set(successJobResponseOnFireLogIDToCacheKey(onFireID), task.JobResponse, cache.DefaultExpiration)
		delete(s.onFireID2Waiter, onFireID)
	} else {
		//如果执行失败，给第一个返回true，让第一个执行
		if len(waitChanSlice) > 0 {
			waitChanSlice[0] <- true
			s.onFireID2Waiter[onFireID] = s.onFireID2Waiter[onFireID][1:]
		}
	}
}

// CheckDuplicateExecuteSuccessJob 检查请求过来的任务是不是已经执行成功了。如果是的话，返回上次的执行成功的记录，
// 而如果上次执行失败了，或者根本没执行过这个任务，返回nil
func (s *DuplicateService) CheckDuplicateExecuteSuccessJob(onFireID uint) *api.RunJobResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ret, ok := s.c.Get(successJobResponseOnFireLogIDToCacheKey(onFireID)); !ok {
		return nil
	} else {
		//延长一下过期时间
		s.c.Set(successJobResponseOnFireLogIDToCacheKey(onFireID), ret, cache.DefaultExpiration)
		return ret.(*api.RunJobResponse)
	}
}

// CheckDuplicateExecute 检查当前的OnFireLog是否正在执行。如果是的话，返回一个chan bool，如果不为nil，
// 调用方需要等待这个chan bool，如果返回true，则表示刚刚执行结束的任务失败了，调用方本次可以执行。如果返回false，
// 说明刚刚执行结束的任务执行成功了，调用方本次不执行。考虑最极端的情况，同时有很多个调用方，等待同一个任务执行，那么
// 其顺序应是按到来顺序排队执行，直到第一个执行成功或者轮到自己执行。
// todo 两个合二为一
func (s *DuplicateService) CheckDuplicateExecute(onFireID uint) chan bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.onFireID2Waiter[onFireID]
	if !ok {
		//没有等待者，直接执行即可
		s.onFireID2Waiter[onFireID] = make([]chan bool, 0)
		return nil
	}

	//有等待者，自己需要排队
	myWaitChan := make(chan bool, 1)
	s.onFireID2Waiter[onFireID] = append(s.onFireID2Waiter[onFireID], myWaitChan)
	return myWaitChan
}

func (s *DuplicateService) OnSendRunJobResponse(response *api.RunJobResponse) {

}

func successJobResponseOnFireLogIDToCacheKey(onFireID uint) string {
	return strconv.Itoa(int(onFireID))
}

var _ ExecuteListener = (*DuplicateService)(nil)
