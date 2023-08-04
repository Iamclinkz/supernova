package service

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"sync"
	"time"
)

type OnFireService struct {
	instanceID        string
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService

	successBuffer []struct {
		ID     uint
		Result string
	}
	successBufferLock sync.Mutex
	successBufferCond *sync.Cond
	stopCh            chan struct{}
	wg                sync.WaitGroup
}

func NewOnFireService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService, instanceID string) *OnFireService {
	successBufferCond := sync.NewCond(&sync.Mutex{})
	return &OnFireService{
		scheduleOperator:  jobOperator,
		statisticsService: statisticsService,
		successBuffer: make([]struct {
			ID     uint
			Result string
		}, 0),
		successBufferLock: sync.Mutex{},
		successBufferCond: successBufferCond,
		instanceID:        instanceID,
		stopCh:            make(chan struct{}),
		wg:                sync.WaitGroup{},
	}
}
func (s *OnFireService) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	return s.scheduleOperator.UpdateOnFireLogExecutorStatus(ctx, onFireLog)
}

func (s *OnFireService) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	return s.scheduleOperator.UpdateOnFireLogFail(ctx, onFireLogID, errorMsg)
}

func (s *OnFireService) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	s.successBufferLock.Lock()
	s.successBuffer = append(s.successBuffer, struct {
		ID     uint
		Result string
	}{ID: onFireLogID, Result: result})
	s.successBufferLock.Unlock()

	// 通知批量更新的goroutine
	s.successBufferCond.Signal()

	return nil
}

func (s *OnFireService) Start() {
	s.batchUpdateOnFireLogsSuccess()
}

func (s *OnFireService) Stop() {
	s.wg.Add(1)
	close(s.stopCh)
	s.wg.Wait()
}

func (s *OnFireService) batchUpdateOnFireLogsSuccess() {
	const (
		batchSize         = 10000
		maxBufferDuration = 1 * time.Second
	)
	klog.Infof("[%v]batchUpdateOnFireLogsSuccess stopped", s.instanceID)
	select {
	case <-s.stopCh:
		klog.Infof("[%v]batchUpdateOnFireLogsSuccess stopped", s.instanceID)
		s.wg.Done()
		return
	default:
	}

	for {
		s.successBufferCond.L.Lock()
		timeout := time.After(maxBufferDuration)
		for len(s.successBuffer) < batchSize {
			select {
			case <-timeout:
				break
			default:
				s.successBufferCond.Wait()
			}
		}
		s.successBufferCond.L.Unlock()

		s.successBufferLock.Lock()
		if len(s.successBuffer) > 0 {
			if err := s.scheduleOperator.UpdateOnFireLogsSuccess(context.TODO(), s.successBuffer); err != nil {
				klog.Errorf("[%v]batchUpdateOnFireLogsSuccess error: %v", s.instanceID, err)
			} else {
				s.successBuffer = s.successBuffer[:0]
			}
		}
		s.successBufferLock.Unlock()
	}
}

func (s *OnFireService) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	return s.scheduleOperator.UpdateOnFireLogStop(ctx, onFireLogID, msg)
}
