package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/angellee177/go-tasks-crud/models"
	services "github.com/angellee177/go-tasks-crud/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	service services.ITaskService
}

func NewTaskHandler(service services.ITaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) AddTask(ctx *gin.Context) {
	var task models.Task

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val, _ := ctx.Get("userId")
	userID, _ := uuid.Parse(val.(string))

	// Pass the Gin context to implements context.Context.
	if err := h.service.CreateTask(ctx, &task, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create task"})
		return
	}

	ctx.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTasks(ctx *gin.Context) {
	// Get userID from middleware
	val, _ := ctx.Get("userId")
	userID, _ := uuid.Parse(val.(string))

	// Extract params
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	author := ctx.Query("author")
	title := ctx.Query("title")
	status := models.TaskStatus(ctx.Query("status")) // Cast string to TaskStatus
	search := ctx.Query("search")

	// Call Service
	tasks, total, err := h.service.GetTasks(
		ctx,
		title,
		author,
		status,
		search,
		limit,
		offset,
		userID,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Simple JSON response
	ctx.JSON(http.StatusOK, gin.H{
		"data":          tasks,
		"total_results": total,
		"total_pages":   int(math.Ceil(float64(total) / float64(limit))),
		"page":          page,
		"limit":         limit,
	})
}

func (h *TaskHandler) GetTaskDetail(ctx *gin.Context) {
	idParam := ctx.Param("id")

	taskID, err := uuid.Parse(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	val, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentUserID, _ := uuid.Parse(val.(string))
	task, err := h.service.GetTaskByID(ctx, taskID, currentUserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// return the data
	ctx.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	val, _ := ctx.Get("userId")
	userID, _ := uuid.Parse(val.(string))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	var input models.Task
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedTask, err := h.service.UpdateTask(ctx, id, &input, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTask)
}

func (h *TaskHandler) DeleteTask(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	val, _ := ctx.Get("userId")
	userID, _ := uuid.Parse(val.(string))

	if err := h.service.DeleteTask(ctx, id, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task Deleted successfully"})
}
