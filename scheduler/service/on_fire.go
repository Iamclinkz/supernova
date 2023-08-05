package service

import (
	"context"
	"supernova/scheduler/model"
	"supernova/scheduler/operator/schedule_operator"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

type OnFireService struct {
	instanceID        string
	scheduleOperator  schedule_operator.Operator
	statisticsService *StatisticsService

	stopCh   chan struct{}
	wg       sync.WaitGroup
	updateCh chan struct {
		ID     uint
		Result string
	}
}

func NewOnFireService(jobOperator schedule_operator.Operator, statisticsService *StatisticsService, instanceID string) *OnFireService {
	return &OnFireService{
		scheduleOperator:  jobOperator,
		statisticsService: statisticsService,
		instanceID:        instanceID,
		stopCh:            make(chan struct{}),
		wg:                sync.WaitGroup{},
		updateCh: make(chan struct {
			ID     uint
			Result string
		}, 10240),
	}
}
func (s *OnFireService) UpdateOnFireLogExecutorStatus(ctx context.Context, onFireLog *model.OnFireLog) error {
	return s.scheduleOperator.UpdateOnFireLogExecutorStatus(ctx, onFireLog)
}

func (s *OnFireService) UpdateOnFireLogFail(ctx context.Context, onFireLogID uint, errorMsg string) error {
	return s.scheduleOperator.UpdateOnFireLogFail(ctx, onFireLogID, errorMsg)
}

func (s *OnFireService) UpdateOnFireLogSuccess(ctx context.Context, onFireLogID uint, result string) error {
	s.updateCh <- struct {
		ID     uint
		Result string
	}{ID: onFireLogID, Result: result}

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
	var (
		batchSize         = 1000
		maxBufferDuration = 1 * time.Second

		buffer = make([]struct {
			ID     uint
			Result string
		}, 0, 1024)
	)
	klog.Infof("[%v]batchUpdateOnFireLogsSuccess start", s.instanceID)

	for {
		timeout := time.NewTimer(maxBufferDuration)

		timeoutCame := false
		for len(buffer) < batchSize {
			select {
			case <-s.stopCh:
				klog.Infof("[%v]batchUpdateOnFireLogsSuccess stop", s.instanceID)
				s.wg.Done()
				return
			case log := <-s.updateCh:
				buffer = append(buffer, log)
			case <-timeout.C:
				timeoutCame = true
			}
			if timeoutCame {
				break
			}
		}

		timeout.Stop()

		if len(buffer) > 0 {
			if err := s.scheduleOperator.UpdateOnFireLogsSuccess(context.TODO(), buffer); err != nil {
				klog.Errorf("[%v]batchUpdateOnFireLogsSuccess error: %v", s.instanceID, err)
			}
			buffer = buffer[:0]
		}
	}
}

func (s *OnFireService) UpdateOnFireLogStop(ctx context.Context, onFireLogID uint, msg string) error {
	return s.scheduleOperator.UpdateOnFireLogStop(ctx, onFireLogID, msg)
}
