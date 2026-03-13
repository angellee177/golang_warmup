package models

import (
	"testing"
	"time"

	"github.com/angellee177/go-tasks-crud/common"
	"github.com/angellee177/go-tasks-crud/config"
	"github.com/stretchr/testify/assert"
)

func TestTask_BeforeSave(t *testing.T) {
	// Initialize the DB using config
	db, err := config.Init()
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}

	// AutoMigrate to ensure the table exists
	db.AutoMigrate(&Task{})

	t.Run("Error - Missing Due Date", func(t *testing.T) {
		task := Task{
			Title:   "Test Task",
			DueDate: common.Date(time.Time{}), // Zero time
		}

		err := db.Create(&task).Error

		assert.Error(t, err)
		assert.Equal(t, "due date is required", err.Error())
	})

	t.Run("Error - Past Due Date", func(t *testing.T) {
		// Set date to yesterday
		yesterday := time.Now().AddDate(0, 0, -1)
		task := Task{
			Title:   "Past Task",
			DueDate: common.Date(yesterday),
		}

		err := db.Create(&task).Error

		assert.Error(t, err)
		assert.Equal(t, "due date cannot be in the past", err.Error())
	})

	t.Run("Success - Valid Due Date", func(t *testing.T) {
		// 1. Create a dummy user first to satisfy the Foreign Key
		user := User{Name: "Task Owner", Email: "owner@example.com"}
		db.Create(&user)
		defer db.Unscoped().Delete(&user) // Cleanup user after

		tomorrow := time.Now().AddDate(0, 0, 1)
		task := Task{
			Title:   "Future Task",
			DueDate: common.Date(tomorrow),
			UserID:  user.ID, // Link the task to the user
		}

		err := db.Create(&task).Error

		assert.NoError(t, err)
		db.Unscoped().Delete(&task)
	})
}
