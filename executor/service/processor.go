package service

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	trace2 "go.opentelemetry.io/otel/trace"
	"supernova/executor/processor"
	"supernova/pkg/api"
	"supernova/pkg/session/trace"
	"sync"
	"time"
)

type ProcessorService struct {
	pcs               map[string]*processor.PC
	processorWorkerCh map[string]chan *ProcessJobParam
	stopCh            chan struct{}
	wg                sync.WaitGroup
	instanceID        string
	oTelConfig        *trace.OTelConfig
	tracer            trace2.Tracer
	goCounter         int
}

func NewProcessorService(pcs map[string]*processor.PC, instanceID string, oTelConfig *trace.OTelConfig) *ProcessorService {
	ret := &ProcessorService{
		pcs:               pcs,
		stopCh:            make(chan struct{}),
		processorWorkerCh: make(map[string]chan *ProcessJobParam, len(pcs)),
		wg:                sync.WaitGroup{},
		instanceID:        instanceID,
	}

	if oTelConfig.EnableTrace {
		ret.tracer = otel.Tracer("ProcessorTracer")
	}

	for glue, _ := range pcs {
		ret.processorWorkerCh[glue] = make(chan *ProcessJobParam, 10240)
	}
	return ret
}

func (s *ProcessorService) Start() {
	for glue, conf := range s.pcs {
		ch := s.processorWorkerCh[glue]
		if ch == nil {
			panic("")
		}
		if conf.Config.Async {
			s.goCounter++
			//异步，只搞一个
			go s.asyncWork(ch, conf)
		} else {
			s.goCounter += conf.Config.MaxWorkerCount
			//同步，开指定个
			for i := 0; i < conf.Config.MaxWorkerCount; i++ {
				go s.syncWork(ch, conf)
			}
		}
	}
}

func (s *ProcessorService) Stop() {
	s.wg.Add(s.goCounter)
	close(s.stopCh)
	s.wg.Wait()
}

// asyncWork 如果glueType是异步的，每个glue只有go程运行本函数
func (s *ProcessorService) asyncWork(waitCh <-chan *ProcessJobParam, processConfig *processor.PC) {
	var (
		glue    = processConfig.Processor.GetGlueType()
		limitCh = make(chan struct{}, processConfig.Config.MaxWorkerCount)
		doTrace = s.oTelConfig.EnableTrace
		p       = processConfig.Processor
	)
	klog.Infof("[%v] ProcessorService work for:%v started", s.instanceID, glue)
	for {
		select {
		case <-s.stopCh:
			klog.Infof("[%v] %v worker stopped", s.instanceID, glue)
			s.wg.Done()
			return
		case jobParam := <-waitCh:
			go func(param *ProcessJobParam) {
				<-limitCh
				klog.Tracef("[%v] %v worker receive:%v", s.instanceID, glue, param)
				var (
					executeSpan trace2.Span
				)
				//如果不异步，自己执行
				if doTrace {
					_, executeSpan = s.tracer.Start(param.traceCtx, "executeTask")
				}

				start := time.Now()
				jobResponse := new(api.RunJobResponse)
				jobResponse.OnFireLogID = param.jobRequest.OnFireLogID
				jobResponse.TraceContext = param.jobRequest.TraceContext
				jobResponse.Result = p.Process(param.jobRequest.Job)
				param.after(param.jobRequest, jobResponse, time.Since(start))
				if doTrace {
					executeSpan.End()
				}
				klog.Tracef("[%v] %v worker finish:%v", s.instanceID, glue, jobResponse)
				limitCh <- struct{}{}
			}(jobParam)
		}
	}
}

// syncWork 如果是同步的，收到任务说明是自己执行
func (s *ProcessorService) syncWork(waitCh <-chan *ProcessJobParam, processConfig *processor.PC) {
	var (
		glue    = processConfig.Processor.GetGlueType()
		doTrace = s.oTelConfig.EnableTrace
		p       = processConfig.Processor
	)
	klog.Infof("[%v] ProcessorService work for:%v started", s.instanceID, glue)
	for {
		select {
		case <-s.stopCh:
			klog.Infof("[%v] %v worker stopped", s.instanceID, glue)
			s.wg.Done()
			return
		case param := <-waitCh:
			klog.Tracef("[%v] %v worker receive:%v", s.instanceID, glue, param)
			var (
				executeSpan trace2.Span
			)
			//如果不异步，自己执行
			if doTrace {
				_, executeSpan = s.tracer.Start(param.traceCtx, "executeTask")
			}

			start := time.Now()
			jobResponse := new(api.RunJobResponse)
			jobResponse.OnFireLogID = param.jobRequest.OnFireLogID
			jobResponse.TraceContext = param.jobRequest.TraceContext
			jobResponse.Result = p.Process(param.jobRequest.Job)
			param.after(param.jobRequest, jobResponse, time.Since(start))
			if doTrace {
				executeSpan.End()
			}
			klog.Tracef("[%v] %v worker finish:%v", s.instanceID, glue, jobResponse)
		}
	}
}

type ProcessJobParam struct {
	traceCtx   context.Context
	jobRequest *api.RunJobRequest
	after      func(request *api.RunJobRequest, response *api.RunJobResponse, executeTime time.Duration)
}

func (s *ProcessorService) ProcessJob(glueType string, param *ProcessJobParam) bool {
	ch := s.processorWorkerCh[glueType]
	if ch == nil {
		return false
	}

	ch <- param
	return true
}
