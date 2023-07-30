package service

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"supernova/pkg/api"
	"supernova/pkg/util"
	"supernova/scheduler/constance"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type ScheduleService struct {
	stopCh                chan struct{}
	statisticsService     *StatisticsService
	jobService            *JobService
	triggerService        *TriggerService
	onFireService         *OnFireService
	executorSelectService *ExecutorRouteService
	executorManageService *ExecutorManageService
	timeWheel             *util.TimeWheel //这里不用cron.Cron，因为触发时间确定，且只触发一次，且距离触发时间较短，所以用时间轮定时器了
	timeWheelTaskCh       chan *model.OnFireLog
	overtimeOnFireLogCh   chan *model.OnFireLog
	jobResponseTaskCh     chan *api.RunJobResponse
	wg                    sync.WaitGroup
	workerCount           int
	enableOTel            bool
	tracer                trace.Tracer
}

func NewScheduleService(statisticsService *StatisticsService,
	jobService *JobService, triggerService *TriggerService, onFireService *OnFireService,
	executorSelectService *ExecutorRouteService,
	workerCount int, executorManageService *ExecutorManageService, enableOTel bool) *ScheduleService {
	tw, _ := util.NewTimeWheel(time.Millisecond*200, 512, util.TickSafeMode())
	ret := &ScheduleService{
		stopCh:                make(chan struct{}),
		statisticsService:     statisticsService,
		jobService:            jobService,
		triggerService:        triggerService,
		onFireService:         onFireService,
		executorSelectService: executorSelectService,
		executorManageService: executorManageService,
		timeWheel:             tw,
		//todo 想一下大小
		timeWheelTaskCh:     make(chan *model.OnFireLog, workerCount*20),
		overtimeOnFireLogCh: make(chan *model.OnFireLog, workerCount*20),
		jobResponseTaskCh:   make(chan *api.RunJobResponse, workerCount*20),
		wg:                  sync.WaitGroup{},
		workerCount:         workerCount,
		enableOTel:          enableOTel,
	}
	if enableOTel {
		ret.tracer = otel.Tracer("ScheduleTracer")
	}
	ret.executorManageService.RegisterReceiveMsgNotifyFunc(ret.onReceiveJobResponse)
	return ret
}

func (s *ScheduleService) Schedule() {
	s.timeWheel.Start()
	s.startWorker()
	go s.checkTimeoutOnFireLogs()
	ticker := time.NewTicker(s.statisticsService.GetScheduleInterval())

	klog.Infof("ScheduleService start Schedule, worker count:%v", s.workerCount)
	for {
		select {
		case <-s.stopCh:
			klog.Infof("schedule go routine stopped")
			ticker.Stop()
			return
		case <-ticker.C:
			onFireLogs, err := s.triggerService.fetchUpdateMarkTrigger()
			if err != nil {
				klog.Errorf("Schedule Error:%v", err)
			} else {
				now := time.Now()
				for _, onFireLog := range onFireLogs {
					currentLog := onFireLog
					s.timeWheel.Add(currentLog.ShouldFireAt.Sub(now), func() {
						s.timeWheelTaskCh <- currentLog
					}, false)
				}
			}
			ticker.Reset(s.statisticsService.GetScheduleInterval())
		}
	}
}

func (s *ScheduleService) Stop() {
	close(s.stopCh)
	s.timeWheel.Stop()
	s.wg.Wait()
	klog.Infof("ScheduleService worker, timeWheel, scheduler stopped")

	//todo 优雅退出？要不要处理完管道里的所有消息？
}

// fire 执行一次任务。retry表示是否是重试
func (s *ScheduleService) fire(onFireLog *model.OnFireLog, retry bool) error {
	klog.Tracef("worker start fire:%+v", onFireLog)
	job, err := s.jobService.FetchJobFromID(context.TODO(), onFireLog.JobID)
	if err != nil {
		if err == schedule_operator.ErrNotFound {
			//job找不到了，说明被删且用户不希望失败的任务再执行了。
			//这种情况下，尝试更新一下这条OnFireLog，标记成执行结束了。不让其他进程再取了
			//更新不成功可能是已经成功了或者其他原因，不需要处理
			_ = s.onFireService.UpdateOnFireLogStop(context.TODO(), onFireLog.ID, "user cancel")
			return nil
		} else {
			//其他情况，可能是网络不通？先不操作了
			return errors.New("fetch db failed:" + err.Error())
		}
	}

	if job.Status == constance.JobStatusDeleted && !retry {
		//如果job已经删除了，那么只有重试才能执行。否则不执行
		return nil
	}

	executorWrapper, err := s.executorSelectService.ChooseJobExecutor(job, onFireLog, retry)
	if err != nil || executorWrapper == nil {
		_ = s.onFireService.UpdateOnFireLogFail(context.TODO(), onFireLog.ID, "No matched executors")
		return err
	}

	//确认fire了，搞一个trace
	var (
		span     trace.Span
		traceCtx context.Context
	)

	if s.enableOTel {
		//如果是第一次执行，或者onFireLog中没有Trace信息，那么我们搞一个Trace信息
		if !retry || onFireLog.TraceContext == "" {
			traceAttrs := []attribute.KeyValue{
				attribute.String("onFireID", strconv.Itoa(int(onFireLog.ID))),
			}
			traceCtx, span = s.tracer.Start(context.Background(), "fire", trace.WithAttributes(traceAttrs...))
			onFireLog.TraceContext = util.TraceCtx2String(traceCtx)
		} else {
			parentCtx := util.String2TraceCtx(onFireLog.TraceContext)
			traceCtx, span = s.tracer.Start(parentCtx, "redo fire")
		}
	}

	//更新db中的onFireLog的状态，以及ExecutorInstance实例信息
	onFireLog.ExecutorInstance = executorWrapper.ServiceData.InstanceId
	onFireLog.Status = constance.OnFireStatusExecuting

	if err = s.onFireService.UpdateOnFireLogExecutorStatus(context.TODO(), onFireLog); err != nil {
		if s.enableOTel {
			span.RecordError(fmt.Errorf("fail to update on fire log executor status:%v", err))
			span.End()
		}
		return errors.New("update on fire log status fail:" + err.Error())
	}

	if err = executorWrapper.Operator.RunJob(model.GenRunJobRequest(onFireLog, job, util.GenTraceContext(traceCtx))); err != nil {
		//如果执行出错，说明是流错误。不扣除RetryCount。等下次再执行
		span.RecordError(fmt.Errorf("fail to run job:%v", err))
		span.End()
		return errors.New("run job fail:" + err.Error())
	}
	return nil
}

func (s *ScheduleService) handleRunJobResponse(response *api.RunJobResponse) {
	_, mySpan := util.NewSpanFromTraceContext("handleRunJobResponse", s.tracer, response.TraceContext)
	defer mySpan.End()

	if !response.Result.Ok {
		//如果执行出现了错误，那么需要扣除RetryCount，并且根据总重试次数，和当前重试次数更新下次的执行时间。
		//这里的思考是这样的：
		//如果收到一个失败，那么一定是执行失败（因为Executor一致性路由只能防止任务重复执行成功任务，而如果失败，肯定是经过了一次执行，然后失败了）
		//所以这里收到一个错误，扣除一次RetryCount肯定是没问题的。即使收到错误的顺序和派发任务的顺序可能会不一样。
		//这里更新失败了，有可能是已经成功了，我们收到了落后的失败消息，也可能已经失败了，或者数据库连接失败。
		//对于这几种情况，已经成功/已经失败不需要处理。数据库连接失败最多也就是多尝试执行一次这个失败的任务。
		//总之不是很严重的问题，先不处理了。
		_ = s.onFireService.UpdateOnFireLogFail(context.TODO(), uint(response.OnFireLogID), response.Result.Err)
		klog.Debugf("Receive from executor, [onFireLog:%v] execute failed, error:%v",
			response.OnFireLogID, response.Result.Err)
		mySpan.RecordError(fmt.Errorf("fail to run job:%v", response.Result))
		return
	}

	//如果执行成功，更新OnFireLog
	//todo 更新失败的话，除了一致性路由保底，这里也需要处理下
	_ = s.onFireService.UpdateOnFireLogSuccess(context.TODO(), uint(response.OnFireLogID), response.Result.Result)
	klog.Tracef("Receive from executor, [onFireLog:%v] execute successes", response.OnFireLogID)
}

func (s *ScheduleService) onReceiveJobResponse(response *api.RunJobResponse) {
	s.jobResponseTaskCh <- response
}

func (s *ScheduleService) startWorker() {
	if s.workerCount <= 0 {
		panic("")
	}

	s.wg.Add(s.workerCount)
	for i := 0; i < s.workerCount; i++ {
		go s.work()
	}
	klog.Infof("ScheduleService worker started, count:%v", s.workerCount)
}

// 执行消息的发送和接收
func (s *ScheduleService) work() {
	for {
		select {
		case onFireJob := <-s.timeWheelTaskCh:
			if fireErr := s.fire(onFireJob, false); fireErr != nil {
				klog.Debugf("onFireNormalJob:%v fire error:%v", onFireJob, fireErr)
			}
		case response := <-s.jobResponseTaskCh:
			s.handleRunJobResponse(response)
		case overtimeOnFireLog := <-s.overtimeOnFireLogCh:
			if fireErr := s.fire(overtimeOnFireLog, true); fireErr != nil {
				klog.Debugf("onFireOvertimeJob:%v fire error:%v", overtimeOnFireLog, fireErr)
			}
		case <-s.stopCh:
			klog.Infof("ScheduleService worker stop working")
			s.wg.Done()
			return
		}
	}
}

// checkTimeoutOnFireLogs 从数据库中捞取超时的OnFireLogs，尝试再次执行
func (s *ScheduleService) checkTimeoutOnFireLogs() {
	ticker := time.NewTicker(s.statisticsService.GetCheckTimeoutOnFireLogsInterval())

	for {
		select {
		case <-s.stopCh:
			klog.Infof("checkTimeoutOnFireLogs go routine stopped")
			ticker.Stop()
			return
		case <-ticker.C:
			onFireLogs, err := s.triggerService.fetchTimeoutAndRefreshOnFireLogs()
			if err != nil {
				klog.Errorf("checkTimeoutOnFireLogs Error:%v", err)
			} else {
				for _, onFireLog := range onFireLogs {
					s.overtimeOnFireLogCh <- onFireLog
				}
			}
			ticker.Reset(s.statisticsService.GetCheckTimeoutOnFireLogsInterval())
		}
	}
}
