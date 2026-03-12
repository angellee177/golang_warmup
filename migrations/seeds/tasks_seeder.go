package seeds

import (
	"log"
	"time"

	"github.com/angellee177/go-tasks-crud/common"
	"github.com/angellee177/go-tasks-crud/models"
	"github.com/brianvoe/gofakeit/v6"
	"gorm.io/gorm"
)

func SeedTask(db *gorm.DB) {
	log.Println("🌱 Seeding database...")

	// Clear table before generate the seeder.
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Task{})

	// Get all user IDs to link tasks
	var users []models.User
	db.Find(&users)

	if len(users) == 0 {
		log.Println("❌ Skipping task seeding: No users found. Seed users first!")
		return
	}

	count := 50
	tasks := make([]models.Task, count)

	for i := 0; i < count; i++ {
		targetUser := users[i%2]

		// Assign the fake tasks into task at index.
		tasks[i] = models.Task{
			Title:       gofakeit.Sentence(3),
			Description: gofakeit.Paragraph(1, 2, 5, " "), // Random text
			Status:      models.StatusTodo,
			DueDate:     common.Date(time.Now().AddDate(0, 0, 7)), // Due in 7 days
			UserID:      targetUser.ID,                            // Link to real user
		}
	}

	// Bulk Insert in single SQL query for 50 tasks.
	result := db.CreateInBatches(tasks, 50)

	if result.Error != nil {
		log.Printf("❌ Seeding failed: %v", result.Error)
		return
	}

	log.Printf("✅ Seeding success: %d tasks added to the database.", result.RowsAffected)
}
