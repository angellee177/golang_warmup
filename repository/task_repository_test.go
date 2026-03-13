package repository

import (
	"context"
	"testing"
	"time"

	"github.com/angellee177/go-tasks-crud/common"
	"github.com/angellee177/go-tasks-crud/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTaskRepository_Integration(t *testing.T) {
	db, cleanup := setupGenericTestContainer(t)
	defer cleanup()

	// Ensure tables exist for Task logic
	db.AutoMigrate(&models.User{}, &models.Task{})

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Setup Create a test User
	user := models.User{ID: uuid.New(), Name: "user1", Email: "user1@example.com"}
	assert.NoError(t, db.Create(&user).Error)

	// Setup create a few tasks
	// to fake the Due Date
	tomorrow := time.Now().Add(24 * time.Hour)
	task1 := models.Task{
		ID:      uuid.New(),
		Title:   "Complete Go CRUD",
		Status:  models.StatusInProgress,
		UserID:  user.ID,
		DueDate: common.Date(tomorrow),
	}
	task2 := models.Task{
		ID:      uuid.New(),
		Title:   "Write Integration Tests",
		Status:  models.StatusTodo,
		UserID:  user.ID,
		DueDate: common.Date(tomorrow),
	}

	// VERIFY CREATE WORKS
	errCreate := db.Create(&task1).Error
	if errCreate != nil {
		t.Fatalf("Failed to create task setup: %v", errCreate)
	}
	db.Create(&task2)

	t.Run("FindByIDAndUser -- Success", func(t *testing.T) {
		found, err := repo.FindByIDAndUser(ctx, task1.ID, user.ID)

		if assert.NoError(t, err) && assert.NotNil(t, found) {
			assert.Equal(t, task1.Title, found.Title)
		}
	})

	t.Run("FindByIDAndUser - Not Found (Wrong User)", func(t *testing.T) {
		wrongUserID := uuid.New()
		_, err := repo.FindByIDAndUser(ctx, task1.ID, wrongUserID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("SearchTasks - Filter by Status", func(t *testing.T) {
		tasks, total, err := repo.SearchTasks(ctx, "", "", models.StatusTodo, "", 10, 0, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Write Integration Tests", tasks[0].Title)
	})

	t.Run("SearchTasks - Filter by Title", func(t *testing.T) {
		// This specifically triggers the 'if title != ""' block
		// Searching for "Complete" should match "Complete Go CRUD"
		tasks, total, err := repo.SearchTasks(ctx, "Complete", "", "", "", 10, 0, user.ID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		if assert.Len(t, tasks, 1) {
			assert.Equal(t, "Complete Go CRUD", tasks[0].Title)
		}
	})

	t.Run("SearchTasks - Filter by Title (Case Insensitive)", func(t *testing.T) {
		// This proves ILIKE is working for the title field
		tasks, total, err := repo.SearchTasks(ctx, "go crud", "", "", "", 10, 0, user.ID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		if assert.Len(t, tasks, 1) {
			assert.Contains(t, tasks[0].Title, "Go CRUD")
		}
	})

	t.Run("SearchTasks - Global Search (ILIKE Title)", func(t *testing.T) {
		// Search for "CRUD" (case insensitive) to match task1
		tasks, total, err := repo.SearchTasks(ctx, "", "", "", "crud", 10, 0, user.ID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		if assert.Len(t, tasks, 1) { // Prevents panic if slice is empty
			assert.Contains(t, tasks[0].Title, "CRUD")
		}
	})

	t.Run("SearchTasks - Join Filter by Author Name", func(t *testing.T) {
		// This tests the .Joins("JOIN users...") logic
		tasks, total, err := repo.SearchTasks(ctx, "", "user1", "", "", 10, 0, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.NotNil(t, tasks[0].User)
		assert.Equal(t, "user1", tasks[0].User.Name)
	})

	t.Run("Inherited Generic Methods", func(t *testing.T) {
		// This verifies that the composition works
		err := repo.Delete(ctx, task2.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, task2.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}
