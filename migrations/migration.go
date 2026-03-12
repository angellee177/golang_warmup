package migrations

import (
	"log"

	"github.com/angellee177/go-tasks-crud/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	log.Println("🧬 Running database migrations...")

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")

	err := db.AutoMigrate(
		&models.Task{},
		&models.User{},
	)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migrations completed successfully.")
}
