package http

import (
	"net/http"
	"strconv"

	"supernova/scheduler/model"
	"supernova/scheduler/service"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService *service.JobService
}

func NewJobHandler(jobService *service.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

func (h *JobHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/job/:id", h.GetJob)
	router.POST("/job", h.AddJob)
	router.DELETE("/job/:id", h.DeleteJob)
	router.GET("/job", h.FindJobByName)
	router.POST("/job/batch", h.AddJobs)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.jobService.FetchJobFromID(c.Request.Context(), uint(jobID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) AddJob(c *gin.Context) {
	var job model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.jobService.AddJob(c.Request.Context(), &job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job added successfully"})
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobService.DeleteJob(c.Request.Context(), uint(jobID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *JobHandler) AddJobs(c *gin.Context) {
	var jobs []*model.Job
	if err := c.ShouldBindJSON(&jobs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.jobService.AddJobs(c.Request.Context(), jobs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job added successfully"})
}

func (h *JobHandler) FindJobByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing name parameter"})
		return
	}

	job, err := h.jobService.FindJobByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}
