package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/v1")
	{
		TasksRoutes(v1, db)
		UserRoutes(v1, db)
	}
}