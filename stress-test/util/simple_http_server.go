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
	FailRate      float32
	ListeningPort int
	TriggerCount  int
}

const ExecutorIDFieldName = "X-ExecutorID"

type SimpleHttpServer struct {
	serveConfig            *SimpleHttpServerConf
	mu                     sync.Mutex
	executorID2CallTime    map[uint]int
	executorID2SuccessTime map[uint]int
}

func NewSimpleHttpServer(config *SimpleHttpServerConf) *SimpleHttpServer {
	return &SimpleHttpServer{
		serveConfig:            config,
		mu:                     sync.Mutex{},
		executorID2CallTime:    make(map[uint]int),
		executorID2SuccessTime: make(map[uint]int),
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
		if param[ExecutorIDFieldName] == "" {
			panic("")
		}

		//log.Printf("Received param: %v", param)

		fail := rand.Float32() < s.serveConfig.FailRate
		if fail {
			c.String(http.StatusInternalServerError, "Error")
		} else {
			c.String(http.StatusOK, "OK")
		}

		s.UpdateServe(param[ExecutorIDFieldName], fail)
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
		if s.executorID2CallTime[uint(id)] != 0 {
			panic(fmt.Sprintf("dup called:%v", uint(id)))
		}

		s.executorID2CallTime[uint(id)]++
		if !failed {
			s.executorID2SuccessTime[uint(id)]++
		}
	}
}

type Result struct {
	ExecutorID2CallTime    map[uint]int
	ExecutorID2SuccessTime map[uint]int
}

func (s *SimpleHttpServer) PrintResult() {
	result := s.GetResult()

	done := 0
	notDone := 0
	for i := 0; i < s.serveConfig.TriggerCount; i++ {
		if result.ExecutorID2SuccessTime[uint(i)] != 1 && result.ExecutorID2SuccessTime[uint(i)] != 0 {
			panic(i)
		} else if result.ExecutorID2CallTime[uint(i)] != 1 {
			notDone++
		} else {
			done++
		}
	}

	log.Printf("done:%v, not done:%v", done, notDone)
}

func (s *SimpleHttpServer) GetResult() *Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := &Result{
		ExecutorID2CallTime:    make(map[uint]int),
		ExecutorID2SuccessTime: make(map[uint]int),
	}

	for k, v := range s.executorID2CallTime {
		result.ExecutorID2CallTime[k] = v
	}
	for k, v := range s.executorID2SuccessTime {
		result.ExecutorID2SuccessTime[k] = v
	}

	return result
}
