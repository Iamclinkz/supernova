package simple_http_server

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SimpleHttpServer struct {
	serveConfig     *SimpleHttpServerInitConf
	CheckResultConf *SimpleHttpServerCheckConf
	mu              sync.Mutex
	calledCount     []int //每个trigger执行次数
	successCount    []int //每个trigger成功执行次数
	shutdownCh      chan struct{}

	firstRequestTime  time.Time
	lastRequestTime   time.Time
	totalRequestCount int64
}

func NewSimpleHttpServer(initConf *SimpleHttpServerInitConf, checkConf *SimpleHttpServerCheckConf) *SimpleHttpServer {
	return &SimpleHttpServer{
		serveConfig:     initConf,
		CheckResultConf: checkConf,
		mu:              sync.Mutex{},
		calledCount:     make([]int, initConf.TriggerCount),
		successCount:    make([]int, initConf.TriggerCount),
		shutdownCh:      make(chan struct{}),
	}
}

func (s *SimpleHttpServer) Start() {
	log.Printf("SimpleHttpServer started on port:%v\n", s.serveConfig.ListeningPort)
	router := gin.New()

	router.POST("/test", s.handleTest)
	router.GET("/view", s.handleView)
	router.GET("/shutdown", s.handleShutdown)
	router.GET("/executor-log", s.handleExecutorLog)
	router.GET("/scheduler-log", s.handleSchedulerLog)

	err := router.Run(fmt.Sprintf("0.0.0.0:%d", s.serveConfig.ListeningPort))
	if err != nil {
		panic("")
	}
}

func (s *SimpleHttpServer) IsFail(onFireLogID uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	fail := false
	if s.serveConfig.SuccessAfterFirstFail && s.calledCount[onFireLogID] != 0 {
		//如果指定了第一次失败，其余成功，并且之前失败过了，那么肯定成功
		fail = false
	} else {
		fail = rand.Float32() < s.serveConfig.FailRate
	}

	s.UpdateServe(onFireLogID, fail)
	return fail
}

func (s *SimpleHttpServer) UpdateServe(onFireLogID uint, failed bool) {
	if s.successCount[onFireLogID] != 0 && !s.serveConfig.AllowDuplicateCalled {
		//如果之前执行成功了，现在又收到一条消息，那么是重复执行了。如果指定不能重复执行，则失败
		log.Fatalf("simple http server duplicate called by triggerID:%v\n", onFireLogID)
	}

	s.calledCount[onFireLogID]++
	if !failed {
		s.successCount[onFireLogID]++
	}
}
