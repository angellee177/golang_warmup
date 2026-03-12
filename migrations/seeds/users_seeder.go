package seeds

import (
	"log"

	"github.com/angellee177/go-tasks-crud/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	log.Println("👤 Seeding users...")

	// Clear existing users (Cascade will handle deleting their tasks)
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.User{})

	// Prepare a common hashed password for all seed users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)

	count := 10
	users := []models.User{
		{
			Name:     "user1",
			Email:    "user1@example.com",
			Password: string(hashedPassword),
		},
		{
			Name:     "user2",
			Email:    "user2@example.com",
			Password: string(hashedPassword),
		},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Printf("❌ User seeding failed: %v", err)
		return
	}

	log.Printf("✅ Seeding success: %d users added.", count)
}
