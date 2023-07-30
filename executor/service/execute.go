package service

import (
	"context"
	"errors"
	"supernova/executor/constance"
	"supernova/pkg/api"
	"supernova/pkg/util"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/cloudwego/kitex/pkg/klog"
)

type ExecuteService struct {
	jobRequestCh      chan *api.RunJobRequest
	jobResponseCh     chan *api.RunJobResponse
	statisticsService *StatisticsService
	processorService  *ProcessorService
	duplicateService  *DuplicateService
	stopCh            chan struct{}
	processorCount    int
	wg                *sync.WaitGroup
	timeWheel         *util.TimeWheel
	executeListeners  []ExecuteListener
	enableOTel        bool
	tracer            trace.Tracer
}

func NewExecuteService(statisticsService *StatisticsService, processorService *ProcessorService,
	duplicateService *DuplicateService, processorCount int, enableOTel bool) *ExecuteService {
	if processorCount <= 0 || processorCount > 512 {
		processorCount = 512
	}

	var err error
	tw, err := util.NewTimeWheel(time.Millisecond*200, 100, util.TickSafeMode())
	if err != nil {
		panic(err)
	}

	ret := &ExecuteService{
		jobRequestCh:      make(chan *api.RunJobRequest, processorCount*20),
		jobResponseCh:     make(chan *api.RunJobResponse, processorCount*20),
		statisticsService: statisticsService,
		processorService:  processorService,
		duplicateService:  duplicateService,
		stopCh:            make(chan struct{}),
		processorCount:    processorCount,
		wg:                &sync.WaitGroup{},
		timeWheel:         tw,
		executeListeners:  make([]ExecuteListener, 0, 2),
		enableOTel:        enableOTel,
	}
	if enableOTel {
		ret.tracer = otel.Tracer("ExecuteTracer")
	}
	ret.RegisterExecuteListener(statisticsService)
	ret.RegisterExecuteListener(duplicateService)
	return ret
}

func (e *ExecuteService) PushJobRequest(jobRequest *api.RunJobRequest) {
	e.jobRequestCh <- jobRequest
}

func (e *ExecuteService) PushJobResponse(response *api.RunJobResponse) {
	e.jobResponseCh <- response
}

func (e *ExecuteService) PopJobResponse() (*api.RunJobResponse, bool) {
	taskRet, ok := <-e.jobResponseCh
	return taskRet, ok
}

// Start 开启很多个go程抢夺job运行
func (e *ExecuteService) Start() {
	for i := 0; i < e.processorCount; i++ {
		go e.work()
	}

	klog.Infof("ExecuteService start working, worker count:%v", e.processorCount)
	e.timeWheel.Start()
}

func (e *ExecuteService) Stop() {
	e.wg.Add(e.processorCount)
	close(e.stopCh)
	e.wg.Wait()
	e.timeWheel.Stop()

	klog.Infof("stop work Job")
}

func (e *ExecuteService) work() {
	for {
		select {
		case <-e.stopCh:
			e.wg.Done()
			return
		case jobRequest := <-e.jobRequestCh:
			var (
				doTrace = len(jobRequest.TraceContext) != 0 && e.enableOTel

				workCtx context.Context

				workSpan             trace.Span
				dupWaitExecuteSpan   trace.Span
				executeSpan          trace.Span
				pushResponseChanSpan trace.Span
			)
			if doTrace {
				workCtx, workSpan = util.NewSpanFromTraceContext("executorWork", e.tracer, jobRequest.TraceContext)
			}

			//防止重复执行已经成功的任务
			resp, myWaitCh := e.duplicateService.CheckDuplicateExecuteSuccessJobAndConcurrentExecute(uint(jobRequest.OnFireLogID))
			if resp != nil {
				if !resp.Result.Ok {
					//todo 删掉
					panic("")
				}
				//之前执行过，且执行成功了，直接返回
				klog.Warnf("receive duplicate execute success job request, OnFireID:%v", jobRequest.OnFireLogID)
				e.jobResponseCh <- resp
				if doTrace {
					workSpan.RecordError(errors.New("duplicate execute success job"))
					workSpan.End()
				}
				continue
			}

			//防止当前并发执行同一个任务
			if myWaitCh != nil {
				klog.Warnf("find concurrent execute job request, onFireID:%v", jobRequest.OnFireLogID)
				if doTrace {
					_, dupWaitExecuteSpan = e.tracer.Start(workCtx, "dupWaitExecute")
				}
				needDo := <-myWaitCh
				if !needDo {
					if doTrace {
						dupWaitExecuteSpan.RecordError(errors.New("previous task executed successfully"))
						dupWaitExecuteSpan.End()
						workSpan.End()
					}
					e.statisticsService.OnNoNeedSendResponse(jobRequest)
					continue
				}
			}

			//给时间轮加一个定时事件，如果超时，那么返回一个失败的response
			task := e.timeWheel.Add(time.Microsecond*time.Duration(jobRequest.Job.ExecutorExecuteTimeoutMs), func() {
				klog.Debugf("on job overtime:%v", jobRequest.OnFireLogID)
				e.jobResponseCh <- &api.RunJobResponse{
					OnFireLogID:  jobRequest.OnFireLogID,
					TraceContext: jobRequest.TraceContext,
					Result: &api.JobResult{
						Ok:     false,
						Err:    constance.ExecuteTimeoutErrMsg,
						Result: "",
					},
				}
			}, false)

			klog.Tracef("worker start handle job, OnFireID:%v", jobRequest.OnFireLogID)
			if doTrace {
				_, executeSpan = e.tracer.Start(workCtx, "executeTask")
			}
			e.notifyOnStartExecute(jobRequest)
			jobResponse := new(api.RunJobResponse)
			jobResponse.OnFireLogID = jobRequest.OnFireLogID
			jobResponse.TraceContext = jobRequest.TraceContext

			processor := e.processorService.GetRegister(jobRequest.Job.GlueType)
			if processor == nil {
				//如果找不到Processor，那么框架填充error
				klog.Errorf("can not find processor for OnFireID:%v, glueType required:%v", jobRequest.OnFireLogID,
					jobRequest.Job.GlueType)
				jobResponse.Result = new(api.JobResult)
				jobResponse.Result.Ok = false
				jobResponse.Result.Err = constance.CanNotFindProcessorErrMsg
			} else {
				//如果有Processor，那么用户自定义的Processor填充error
				jobResponse.Result = processor.Process(jobRequest.Job)
			}
			e.notifyOnFinishExecute(jobRequest, jobResponse)

			ok := e.timeWheel.Remove(task)
			//这里如果不ok，说明时间轮定时器中的内容没了，说明本次执行超时，定时器触发，且一定已经返回了一个超时错误。
			//这种情况下，如果本次虽然超时，但是任务执行成功了，则扔回去一个成功回复。而如果执行失败，就不再扔回去失败回复了。
			//反正已经是扣除失败次数了
			if ok {
				if doTrace {
					executeSpan.End()
					_, pushResponseChanSpan = e.tracer.Start(workCtx, "pushResponseChan")
				}
				e.jobResponseCh <- jobResponse
				if doTrace {
					pushResponseChanSpan.End()
				}
			} else {
				if doTrace {
					executeSpan.RecordError(errors.New("execute overtime"))
					executeSpan.End()
				}
				if jobResponse.Result.Ok {
					e.statisticsService.OnOverTimeTaskExecuteSuccess(jobRequest, jobResponse)
					if doTrace {
						_, pushResponseChanSpan = e.tracer.Start(workCtx, "pushResponseChan")
						e.jobResponseCh <- jobResponse
						pushResponseChanSpan.End()
					}
				}
			}

			klog.Tracef("worker execute job finished:%+v", jobResponse)
			if doTrace {
				workSpan.End()
			}
		}
	}
}

type ExecuteListener interface {
	OnStartExecute(request *api.RunJobRequest)                                //开始执行某个任务
	OnFinishExecute(request *api.RunJobRequest, response *api.RunJobResponse) //执行结束某个任务
}

func (e *ExecuteService) RegisterExecuteListener(listener ExecuteListener) {
	e.executeListeners = append(e.executeListeners, listener)
}

func (e *ExecuteService) notifyOnStartExecute(request *api.RunJobRequest) {
	for _, listener := range e.executeListeners {
		listener.OnStartExecute(request)
	}
}

func (e *ExecuteService) notifyOnFinishExecute(request *api.RunJobRequest, response *api.RunJobResponse) {
	for _, listener := range e.executeListeners {
		listener.OnFinishExecute(request, response)
	}
}
