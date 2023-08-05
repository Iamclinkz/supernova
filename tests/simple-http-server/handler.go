package simple_http_server

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"supernova/tests/util"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *SimpleHttpServer) handleTest(c *gin.Context) {
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
	s.mu.Lock()
	if s.totalRequestCount == 0 {
		s.firstRequestTime = time.Now()
	}
	s.lastRequestTime = time.Now()
	s.totalRequestCount++
	s.mu.Unlock()
}

func (s *SimpleHttpServer) handleView(c *gin.Context) {
	result := s.GetResult()

	// 创建一个HTML模板
	tmpl := template.Must(template.New("view").Parse(mainViewTemplate))

	// 将当前的配置和任务执行情况传递给模板，并执行模板
	err := tmpl.Execute(c.Writer, gin.H{
		"ServeConfig": s.serveConfig,
		"Result":      result,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		c.String(http.StatusInternalServerError, "Error")
	}
}

func (s *SimpleHttpServer) handleShutdown(c *gin.Context) {
	log.Println("Shutting down SimpleHttpServer...")
	c.String(http.StatusOK, "Shutting down...")
	close(s.shutdownCh)
}

func (s *SimpleHttpServer) handleExecutorLog(c *gin.Context) {
	logContent, err := ioutil.ReadFile(util.GracefulStoppedExecutorLogPath)
	if err != nil {
		log.Printf("Error reading executor log: %v", err)
		c.String(http.StatusInternalServerError, "Error reading executor log")
	} else {
		c.String(http.StatusOK, string(logContent))
	}
}

func (s *SimpleHttpServer) handleSchedulerLog(c *gin.Context) {
	logContent, err := ioutil.ReadFile(util.KilledSchedulerLogPath)
	if err != nil {
		log.Printf("Error reading scheduler log: %v", err)
		c.String(http.StatusInternalServerError, "Error reading scheduler log")
	} else {
		c.String(http.StatusOK, string(logContent))
	}
}
