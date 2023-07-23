package util

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

//监听/test的POST方法，记录并统计收到的http请求中的triggerID字段的值，根据fail字段随机失败。
//如果失败，返回500，Error，如果不失败，返回200，OK。无论失败与否都会记录数据。

type SimpleHttpServerConf struct {
	FailRate             float32 //随机失败率是多少。如果是0，则不失败
	ListeningPort        int     //监听的端口
	TriggerCount         int     //应该来的trigger
	AllowDuplicateCalled bool    //是否允许一个trigger执行多次？（需要结合 “最多一次“ 和 ”最少一次“ 语义指定）
}

const TriggerIDFieldName = "X-ExecutorID"

type SimpleHttpServer struct {
	serveConfig  *SimpleHttpServerConf
	mu           sync.Mutex
	calledCount  []int //每个trigger执行次数
	successCount []int //每个trigger成功执行次数
}

func NewSimpleHttpServer(config *SimpleHttpServerConf) *SimpleHttpServer {
	return &SimpleHttpServer{
		serveConfig:  config,
		mu:           sync.Mutex{},
		calledCount:  make([]int, config.TriggerCount),
		successCount: make([]int, config.TriggerCount),
	}
}

func (s *SimpleHttpServer) Start() {
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		paramJSON := c.GetHeader("param")
		if paramJSON == "" {
			panic("")
		}
		var param map[string]string
		err := json.Unmarshal([]byte(paramJSON), &param)
		if err != nil {
			panic("")
		}
		if param[TriggerIDFieldName] == "" {
			panic("")
		}

		fail := rand.Float32() < s.serveConfig.FailRate
		if fail {
			c.String(http.StatusInternalServerError, "Error")
		} else {
			c.String(http.StatusOK, "OK")
		}

		s.UpdateServe(param[TriggerIDFieldName], fail)
	})

	err := router.Run(fmt.Sprintf(":%d", s.serveConfig.ListeningPort))
	if err != nil {
		panic("")
	}
}

func (s *SimpleHttpServer) UpdateServe(onFireLogID string, failed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, err := strconv.Atoi(onFireLogID); err != nil {
		panic(err)
	} else {
		if s.successCount[uint(id)] != 0 && !s.serveConfig.AllowDuplicateCalled {
			//如果之前执行成功了，现在又收到一条消息，那么是重复执行了。如果指定不能重复执行，则失败
			log.Fatalf("simple http server duplicate called by triggerID:%v\n", uint(id))
		}

		s.calledCount[uint(id)]++
		if !failed {
			s.successCount[uint(id)]++
		}
	}
}

type Result struct {
	SuccessCount       int     //成功执行的任务个数
	HaveNotCalledCount int     //一次都没执行过的任务个数
	CalledButFailCount int     //执行过，但是最终结果是失败的任务个数
	CalledTotal        int     //总计被执行的次数
	UncalledTriggers   []uint  //一次都没有执行过的trigger
	FailedTriggers     []uint  //到最后，仍然没有执行成功的trigger
	FailTriggerRate    float32 //失败的trigger的个数
}

func (r Result) String() string {
	return fmt.Sprintf("Result:\n"+
		"  SuccessCount: %d\n"+
		"  HaveNotCalledCount: %d\n"+
		"  CalledButFailCount: %d\n"+
		"  CalledTotal: %d\n"+
		"  UncalledTriggers: %v\n"+
		"  FailedTriggers: %v",
		r.SuccessCount,
		r.HaveNotCalledCount,
		r.CalledButFailCount,
		r.CalledTotal,
		r.UncalledTriggers,
		r.FailedTriggers,
	)
}

type CheckResultConf struct {
	AllSuccess                    bool    //一定要都成功
	NoUncalledTriggers            bool    //一定不能有没有执行过的trigger
	FailTriggerRateNotGreaterThan float32 //失败率不能高于
}

func (s *SimpleHttpServer) CheckResult(crc *CheckResultConf) {
	result := s.GetResult()

	if crc.AllSuccess && result.SuccessCount != s.serveConfig.TriggerCount {
		//检查失败
		log.Fatalf("not all triggers successed:%s", result.String())
	}

	if crc.NoUncalledTriggers && len(result.UncalledTriggers) != 0 {
		//有任何一个trigger一次都没有被执行过
		log.Fatalf("not all triggers called:%s", result.String())
	}

	if crc.FailTriggerRateNotGreaterThan < result.FailTriggerRate {
		//失败率检查
		log.Fatalf("high fail rate:%s", result.String())
	}
}

func (s *SimpleHttpServer) GetResult() *Result {
	result := &Result{
		UncalledTriggers: make([]uint, 0),
		FailedTriggers:   make([]uint, 0),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for triggerID := 0; triggerID < s.serveConfig.TriggerCount; triggerID++ {
		successCount := s.successCount[triggerID]
		calledTimes := s.calledCount[triggerID]

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
