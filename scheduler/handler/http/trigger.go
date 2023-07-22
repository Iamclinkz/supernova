package http

import (
	"net/http"
	"strconv"

	"supernova/scheduler/model"
	"supernova/scheduler/service"

	"github.com/gin-gonic/gin"
)

type TriggerHandler struct {
	triggerService *service.TriggerService
}

func NewTriggerHandler(triggerService *service.TriggerService) *TriggerHandler {
	return &TriggerHandler{
		triggerService: triggerService,
	}
}

func (h *TriggerHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/trigger/:id", h.GetTrigger)
	router.POST("/trigger", h.AddTrigger)
	router.DELETE("/trigger/:id", h.DeleteTrigger)
	router.GET("/trigger", h.GetTriggerByName)
	router.POST("/trigger/batch", h.AddTriggers)
}

func (h *TriggerHandler) GetTrigger(c *gin.Context) {
	triggerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger ID"})
		return
	}

	trigger, err := h.triggerService.FetchTriggerFromID(uint(triggerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trigger)
}

func (h *TriggerHandler) AddTrigger(c *gin.Context) {
	var trigger model.Trigger
	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.triggerService.AddTrigger(&trigger); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trigger added successfully"})
}

func (h *TriggerHandler) DeleteTrigger(c *gin.Context) {
	triggerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger ID"})
		return
	}

	if err := h.triggerService.DeleteTrigger(uint(triggerID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trigger deleted successfully"})
}

func (h *TriggerHandler) AddTriggers(c *gin.Context) {
	var triggers []*model.Trigger
	if err := c.ShouldBindJSON(&triggers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.triggerService.AddTriggers(triggers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Triggers added successfully"})
}

func (h *TriggerHandler) GetTriggerByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing name parameter"})
		return
	}

	trigger, err := h.triggerService.FindTriggerByName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trigger)
}
