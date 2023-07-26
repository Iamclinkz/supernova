package simple_http_server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

type SimpleHttpServer struct {
	serveConfig     *SimpleHttpServerInitConf
	CheckResultConf *SimpleHttpServerCheckConf
	mu              sync.Mutex
	calledCount     []int //每个trigger执行次数
	successCount    []int //每个trigger成功执行次数
}

func NewSimpleHttpServer(initConf *SimpleHttpServerInitConf, checkConf *SimpleHttpServerCheckConf) *SimpleHttpServer {
	return &SimpleHttpServer{
		serveConfig:     initConf,
		CheckResultConf: checkConf,
		mu:              sync.Mutex{},
		calledCount:     make([]int, initConf.TriggerCount),
		successCount:    make([]int, initConf.TriggerCount),
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

		var onFireID uint
		if intID, err := strconv.Atoi(param[TriggerIDFieldName]); err != nil {
			panic(err)
		} else {
			onFireID = uint(intID)
		}

		fail := s.IsFail(onFireID)
		if fail {
			c.String(http.StatusInternalServerError, "Error")
		} else {
			c.String(http.StatusOK, "OK")
		}
	})

	err := router.Run(fmt.Sprintf(":%d", s.serveConfig.ListeningPort))
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
