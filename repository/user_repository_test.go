package repository

import (
	"context"
	"testing"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserRepository_Integration(t *testing.T) {
	// Re-use the setup logic from generic_repo_test
	db, cleanup := setupGenericTestContainer(t)
	defer cleanup()

	// Ensure User table exists
	db.AutoMigrate(&models.User{})

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Setup: Test data
	testUser := models.User{
		ID:    uuid.New(),
		Name:  "Gopher",
		Email: "gopher@golang.org",
	}
	assert.NoError(t, db.Create(&testUser).Error)

	t.Run("FindByEmail - Success", func(t *testing.T) {
		// This specifically tests custom method
		found, err := repo.FindByEmail(ctx, "gopher@golang.org")

		assert.NoError(t, err)
		if assert.NotNil(t, found) {
			assert.Equal(t, testUser.ID, found.ID)
			assert.Equal(t, "Gopher", found.Name)
		}
	})

	t.Run("FindByEmail - Not Found", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nonexistent@example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Inherited Generic Create - Success", func(t *testing.T) {
		// Verifies that IGenericRepository[models.User] composition works
		newUser := &models.User{
			ID:    uuid.New(),
			Name:  "Integration Tester",
			Email: "tester@example.com",
		}
		err := repo.Create(ctx, newUser)
		assert.NoError(t, err)

		// Verify it actually hit the DB
		var count int64
		db.Model(&models.User{}).Where("email = ?", "tester@example.com").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error Path - Duplicate Email (Constraint)", func(t *testing.T) {
		// This verifies the DB unique constraint is handled by the repo
		duplicate := &models.User{
			ID:    uuid.New(),
			Name:  "Clone",
			Email: "gopher@golang.org", // Already exists from setup
		}
		err := repo.Create(ctx, duplicate)
		assert.Error(t, err)
	})
}
