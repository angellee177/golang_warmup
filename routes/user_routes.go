package routes

import (
	"log"
	"os"

	"github.com/angellee177/go-tasks-crud/handlers"
	"github.com/angellee177/go-tasks-crud/middleware"
	"github.com/angellee177/go-tasks-crud/repository"
	services "github.com/angellee177/go-tasks-crud/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	// Initialize Repo
	repo := repository.NewUserRepository(db)

	// Initialize Service (Inject Repo)
	service := services.NewUserService(repo)

	// Initialize Handler (Inject Service)
	h := handlers.NewUserHandler(service)

	// Get the Secret from env variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
	}

	userGroup := rg.Group("/users")
	{
		// Public Auth Routes
		userGroup.POST("/register", h.Register)
		userGroup.POST("/login", h.Login)

		// Protected Routes
		protected := userGroup.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtSecret)) // Check JWT here
		{
			protected.GET("/", h.GetUsers)
			protected.GET("/my-profile", h.GetProfile)
			protected.GET("/:id", h.GetUserById)
		}
	}
}
