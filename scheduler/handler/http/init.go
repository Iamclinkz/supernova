package http

import (
	"github.com/gin-gonic/gin"
	"supernova/scheduler/app"
)

func InitHttpHandler(scheduler *app.Scheduler) *gin.Engine {
	// 创建 JobHandler 和 TriggerHandler 实例
	jobHandler := NewJobHandler(scheduler.GetJobService())
	triggerHandler := NewTriggerHandler(scheduler.GetTriggerService())

	// 初始化 Gin 路由
	router := gin.Default()

	// 注册 Job 和 Trigger 路由
	jobRouter := router.Group("/")
	jobHandler.RegisterRoutes(jobRouter)

	triggerRouter := router.Group("/")
	triggerHandler.RegisterRoutes(triggerRouter)

	return router
}
