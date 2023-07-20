package service

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/executor/constance"
	"supernova/pkg/api"
	"supernova/pkg/util"
	"sync"
	"time"
)

type ExecuteService struct {
	taskCh            chan *Task
	jobResponseCh     chan *api.RunJobResponse
	statisticsService *StatisticsService
	processorService  *ProcessorService
	duplicateService  *DuplicateService
	stopCh            chan struct{}
	processorCount    int
	wg                *sync.WaitGroup
	timeWheel         *util.TimeWheel
	executeListeners  []ExecuteListener
}

func NewExecuteService(statisticsService *StatisticsService, processorService *ProcessorService,
	duplicateService *DuplicateService, processorCount int) *ExecuteService {
	if processorCount <= 0 || processorCount > 2000 {
		processorCount = 2000
	}

	tw, _ := util.NewTimeWheel(time.Microsecond*20, 100, util.TickSafeMode())
	ret := &ExecuteService{
		taskCh:            make(chan *Task, processorCount*2),
		jobResponseCh:     make(chan *api.RunJobResponse, processorCount*2),
		statisticsService: statisticsService,
		processorService:  processorService,
		duplicateService:  duplicateService,
		stopCh:            make(chan struct{}),
		processorCount:    processorCount,
		wg:                &sync.WaitGroup{},
		timeWheel:         tw,
		executeListeners:  make([]ExecuteListener, 2),
	}
	ret.RegisterExecuteListener(statisticsService)
	ret.RegisterExecuteListener(duplicateService)
	return ret
}

type Task struct {
	TimeWheelTask *util.Task
	JobRequest    *api.RunJobRequest
	JobResponse   *api.RunJobResponse
}

func (e *ExecuteService) PushJobRequest(jobRequest *api.RunJobRequest) {
	if resp := e.duplicateService.CheckDuplicateExecute(uint(jobRequest.OnFireLogID)); resp != nil && resp.Result.Ok {
		//通过duplicateService防重，即同一个ExecutorLogID的任务，如果执行过，且执行成功了，
		//那么防重不执行，直接从缓存中取出*api.RunJobResponse
		e.jobResponseCh <- resp
		return
	}

	//扔到时间轮里面，如果超时，那么返回一个失败的response
	task := e.timeWheel.Add(time.Microsecond*time.Duration(jobRequest.Job.ExecutorExecuteTimeoutMs), func() {
		e.jobResponseCh <- &api.RunJobResponse{
			OnFireLogID: jobRequest.OnFireLogID,
			Result: &api.JobResult{
				Ok:     false,
				Err:    constance.ExecuteTimeoutErrMsg,
				Result: "",
			},
		}
	}, false)

	e.taskCh <- &Task{
		TimeWheelTask: task,
		JobRequest:    jobRequest,
		JobResponse:   nil,
	}
}

func (e *ExecuteService) PushJobResponse(response *api.RunJobResponse) {
	e.jobResponseCh <- response
}

func (e *ExecuteService) PopJobResponse() (*api.RunJobResponse, bool) {
	taskRet, ok := <-e.jobResponseCh
	return taskRet, ok
}

func (e *ExecuteService) OnTaskFinish(task *Task) {
	e.timeWheel.Remove(task.TimeWheelTask)
	e.jobResponseCh <- task.JobResponse
}

// Start 开启很多个go程抢夺job运行
func (e *ExecuteService) Start() {
	for i := 0; i < e.processorCount; i++ {
		go e.worker()
	}
	e.timeWheel.Start()

	klog.Infof("start to handle Job, go routine count:%v", e.processorCount)
}

func (e *ExecuteService) Stop() {
	e.wg.Add(e.processorCount)
	close(e.stopCh)
	e.wg.Wait()
	e.timeWheel.Stop()

	klog.Infof("stop work Job")
}

func (e *ExecuteService) worker() {
	for {
		select {
		case <-e.stopCh:
			e.wg.Done()
			return
		case task := <-e.taskCh:
			task.JobResponse = new(api.RunJobResponse)
			processor := e.processorService.GetRegister(task.JobRequest.Job.GlueType)
			if processor == nil {
				task.JobResponse.Result = new(api.JobResult)
				task.JobResponse.Result.Ok = false
				task.JobResponse.Result.Err = constance.CanNotFindProcessorErrMsg
			} else {
				task.JobResponse.Result = processor.Process(task.JobRequest.Job)
			}

			e.OnTaskFinish(task)
		}
	}
}

type ExecuteListener interface {
	OnReceiveRunJobRequest(request *api.RunJobRequest)
	OnStartExecute(task *Task)
	OnDuplicateOnFireLogID(request *api.RunJobRequest)
	OnFinishExecute(response *api.RunJobResponse)
}

func (e *ExecuteService) RegisterExecuteListener(listener ExecuteListener) {
	e.executeListeners = append(e.executeListeners, listener)
}

func (e *ExecuteService) OnReceiveRunJobRequest(request *api.RunJobRequest) {
	for _, listener := range e.executeListeners {
		listener.OnReceiveRunJobRequest(request)
	}
}

func (e *ExecuteService) OnStartExecute(task *Task) {
	for _, listener := range e.executeListeners {
		listener.OnStartExecute(task)
	}
}

func (e *ExecuteService) OnDuplicateOnFireLogID(request *api.RunJobRequest) {
	for _, listener := range e.executeListeners {
		listener.OnDuplicateOnFireLogID(request)
	}
}

func (e *ExecuteService) OnFinishExecute(response *api.RunJobResponse) {
	for _, listener := range e.executeListeners {
		listener.OnFinishExecute(response)
	}
}

var _ ExecuteListener = (*ExecuteService)(nil)
