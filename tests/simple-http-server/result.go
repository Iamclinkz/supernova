package simple_http_server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Result struct {
	SuccessCount         int    //成功执行的任务个数
	HaveNotCalledCount   int    //一次都没执行过的任务个数
	CalledButFailCount   int    //执行过，但是最终结果是失败的任务个数
	CalledTotal          int    //总计被执行的次数
	UncalledTriggers     []uint //一次都没有执行过的trigger
	FailedTriggers       []uint //到最后，仍然没有执行成功的trigger
	CalledTwiceOrMore    []uint
	FailTriggerRate      float32 //失败的trigger的个数
	FirstRequestTime     time.Time
	LastRequestTime      time.Time
	AvgRequestsPerSecond float64
}

func (r *Result) String() string {
	return fmt.Sprintf("Result:\n"+
		"  SuccessCount: %d\n"+
		"  HaveNotCalledCount: %d\n"+
		"  CalledButFailCount: %d\n"+
		"  CalledTotal: %d\n"+
		"  UncalledTriggers: %v\n"+
		"  FailedTriggers: %v\n"+
		"  FailTriggerRate: %v\n"+
		"  CalledTwiceOrMore: %v\n",
		r.SuccessCount,
		r.HaveNotCalledCount,
		r.CalledButFailCount,
		r.CalledTotal,
		r.UncalledTriggers,
		r.FailedTriggers,
		r.FailTriggerRate,
		r.CalledTwiceOrMore,
	)
}

// CheckResult 检查结果，如果有不满足的，则error != nil
func (s *SimpleHttpServer) CheckResult() (error, *Result) {
	result := s.GetResult()

	if s.CheckResultConf.AllSuccess && result.SuccessCount != s.serveConfig.TriggerCount {
		//检查失败
		return fmt.Errorf("not all triggers successed"), result
	}

	if s.CheckResultConf.NoUncalledTriggers && len(result.UncalledTriggers) != 0 {
		//有任何一个trigger一次都没有被执行过
		return fmt.Errorf("not all triggers called"), result
	}

	if s.CheckResultConf.FailTriggerRateNotGreaterThan < result.FailTriggerRate {
		//失败率检查
		return fmt.Errorf("high fail rate"), result
	}

	return nil, result
}

func (s *SimpleHttpServer) GetResult() *Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	elapsedTime := s.lastRequestTime.Sub(s.firstRequestTime).Seconds()
	avgRequestsPerSecond := float64(s.totalRequestCount) / elapsedTime
	result := &Result{
		UncalledTriggers:     make([]uint, 0),
		FailedTriggers:       make([]uint, 0),
		FirstRequestTime:     s.firstRequestTime,
		LastRequestTime:      s.lastRequestTime,
		AvgRequestsPerSecond: avgRequestsPerSecond,
	}

	for triggerID := 0; triggerID < s.serveConfig.TriggerCount; triggerID++ {
		successCount := s.successCount[triggerID]
		calledTimes := s.calledCount[triggerID]

		if calledTimes >= 2 {
			result.CalledTwiceOrMore = append(result.CalledTwiceOrMore, uint(triggerID))
		}
		//总calledTimes
		result.CalledTotal += calledTimes

		if successCount != 0 {
			//成功
			result.SuccessCount++
		} else if calledTimes != 0 {
			//尝试过called，但是最终失败了
			result.CalledButFailCount++
			result.FailedTriggers = append(result.FailedTriggers, uint(triggerID))
		} else {
			//一次都没called过
			result.HaveNotCalledCount++
			result.UncalledTriggers = append(result.UncalledTriggers, uint(triggerID))
			result.FailedTriggers = append(result.FailedTriggers, uint(triggerID))
		}
	}

	result.FailTriggerRate = 1 - float32(result.SuccessCount)/float32(s.serveConfig.TriggerCount)
	return result
}

// WaitResult return ok
func (s *SimpleHttpServer) WaitResult(maxWaitTime time.Duration, exit bool) bool {
	timeout := time.After(maxWaitTime)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-timeout:
			err, result := s.CheckResult()
			if err != nil {
				if exit {
					log.Fatalf("CheckResult failed: %v,\n result:%v\n", err, result)
				} else {
					log.Printf("CheckResult failed: %v,\n result:%v\n", err, result)
					return false
				}
			} else {
				log.Printf("result ok:\n%v\n", result)
				return true
			}
		case <-signalCh:
			return true
		default:
			err, result := s.CheckResult()
			if err == nil {
				log.Printf("result ok:\n%v\n", result)
				return true
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *SimpleHttpServer) WaitStop() {
	<-s.shutdownCh
}
