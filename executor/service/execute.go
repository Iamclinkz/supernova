package service

import (
	"supernova/executor/constance"
	"supernova/pkg/api"
	"supernova/pkg/util"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
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
	taskIDCounter     int64 //用于分配task的ID。防止超时任务重复发送执行失败消息
}

func NewExecuteService(statisticsService *StatisticsService, processorService *ProcessorService,
	duplicateService *DuplicateService, processorCount int) *ExecuteService {
	if processorCount <= 0 || processorCount > 2000 {
		processorCount = 2000
	}

	var err error
	tw, err := util.NewTimeWheel(time.Millisecond*20, 100, util.TickSafeMode())
	if err != nil {
		panic(err)
	}

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
	klog.Tracef("receive execute jobRequest, OnFireID:%v", jobRequest.OnFireLogID)
	if resp := e.duplicateService.CheckDuplicateExecuteSuccessJob(uint(jobRequest.OnFireLogID)); resp != nil && resp.Result.Ok {
		//通过duplicateService防重，即同一个ExecutorLogID的任务，如果执行过，且执行成功了，
		//那么防重不执行，直接从缓存中取出*api.RunJobResponse
		klog.Warnf("receive duplicate execute success job request, OnFireID:%v", jobRequest.OnFireLogID)
		e.jobResponseCh <- resp
		return
	}

	//扔到时间轮里面，如果超时，那么返回一个失败的response
	task := e.timeWheel.Add(time.Microsecond*time.Duration(jobRequest.Job.ExecutorExecuteTimeoutMs), func() {
		klog.Warnf("on job overtime:%v", jobRequest.OnFireLogID)
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
	ok := e.timeWheel.Remove(task.TimeWheelTask)
	//这里如果不ok，说明时间轮定时器中的内容没了，只能说明本次执行超时，定时器触发，返回了一个超时错误。
	//这种情况下，如果本次虽然超时，但是任务执行成功了，扔回去一个成功回复。而如果执行失败，就不再扔回去失败回复了。
	//反正已经是扣除失败次数了
	if ok || (!ok && task.JobResponse.Result.Ok) {
		e.jobResponseCh <- task.JobResponse
	}
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
			klog.Tracef("worker start handle job, OnFireID:%v", task.JobRequest.OnFireLogID)
			task.JobResponse = new(api.RunJobResponse)
			processor := e.processorService.GetRegister(task.JobRequest.Job.GlueType)
			task.JobResponse.OnFireLogID = task.JobRequest.OnFireLogID
			if processor == nil {
				//如果找不到Processor，那么框架填充error
				klog.Errorf("can not find processor for OnFireID:%v, glueType required:%v", task.JobRequest.OnFireLogID,
					task.JobRequest.Job.GlueType)
				task.JobResponse.Result = new(api.JobResult)
				task.JobResponse.Result.Ok = false
				task.JobResponse.Result.Err = constance.CanNotFindProcessorErrMsg
			} else {
				//如果有Processor，那么用户自定义的Processor填充error
				task.JobResponse.Result = processor.Process(task.JobRequest.Job)
			}

			klog.Tracef("worker execute task finished:%+v", task)
			e.OnTaskFinish(task)
		}
	}
}

type ExecuteListener interface {
	OnReceiveRunJobRequest(request *api.RunJobRequest) //接收到任务请求
	OnStartExecute(task *Task)                         //开始执行某个任务
	OnFinishExecute(task *Task)                        //执行结束某个任务
	OnSendRunJobResponse(response *api.RunJobResponse) //准备发送任务回复
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

func (e *ExecuteService) OnFinishExecute(task *Task) {
	for _, listener := range e.executeListeners {
		listener.OnFinishExecute(task)
	}
}

func (e *ExecuteService) OnSendRunJobResponse(response *api.RunJobResponse) {
	for _, listener := range e.executeListeners {
		listener.OnSendRunJobResponse(response)
	}
}

var _ ExecuteListener = (*ExecuteService)(nil)
