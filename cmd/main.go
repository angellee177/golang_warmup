package main

import (
	"log"
	"os"

	"github.com/angellee177/go-tasks-crud/config"
	"github.com/angellee177/go-tasks-crud/migrations"
	"github.com/angellee177/go-tasks-crud/migrations/seeds"
	"github.com/angellee177/go-tasks-crud/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	route := gin.Default()

	// Register Routes
	routes.SetupRoutes(route, db)

	// Health check
	route.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "Ping!",
			"db_status": "connected",
		})
	})

	return route
}

func main() {
	// load the .env here for PORT variable
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// init the DB
	handler := config.Init()

	// Run migration and seeder
	migrations.RunMigrations(handler)
	seeds.Run(handler)

	r := SetupRouter(handler)

	// Setup Gin Router
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port fallback
	}

	log.Printf("✅ Database Connection established: %p", handler)
	log.Printf("🚀 Server is starting on port %s...", port)

	r.Run(":" + port)
}
