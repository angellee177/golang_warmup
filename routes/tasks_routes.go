package routes

import (
	"os"

	"github.com/angellee177/go-tasks-crud/handlers"
	"github.com/angellee177/go-tasks-crud/middleware"
	"github.com/angellee177/go-tasks-crud/repository"
	services "github.com/angellee177/go-tasks-crud/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TasksRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	// Init Repo
	repo := repository.NewTaskRepository(db)

	// Init service (Inject Repo)
	service := services.NewTaskService(repo)

	// Init handlers (Inject Service)
	h := handlers.NewTaskHandler(service)

	// Get JWT Secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")

	tasks := rg.Group("/tasks")

	// Apply the AuthMiddleware to all routes in this group
	tasks.Use(middleware.AuthMiddleware(jwtSecret))
	{
		tasks.POST("/", h.AddTask)
		tasks.GET("/", h.GetTasks)
		tasks.GET("/:id", h.GetTaskDetail)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)
	}
}
