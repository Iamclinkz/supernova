package http

import (
	"supernova/scheduler/app"

	"github.com/gin-gonic/gin"
)

func InitHttpHandler(scheduler *app.Scheduler) *gin.Engine {
	jobHandler := NewJobHandler(scheduler.GetJobService())
	triggerHandler := NewTriggerHandler(scheduler.GetTriggerService())

	router := gin.Default()

	//todo 只是debug使用
	//router.Use(middleware.PrintGinHeader)

	jobRouter := router.Group("/")
	jobHandler.RegisterRoutes(jobRouter)

	triggerRouter := router.Group("/")
	triggerHandler.RegisterRoutes(triggerRouter)

	return router
}
