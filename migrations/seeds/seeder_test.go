package seeds

import (
	"testing"

	"github.com/angellee177/go-tasks-crud/config"
	"github.com/angellee177/go-tasks-crud/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupPostgresDB uses your existing config logic
func setupPostgresDB(t *testing.T) *gorm.DB {
	db, err := config.Init()
	if err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}

	// Ensure clean schema for testing
	err = db.AutoMigrate(&models.User{}, &models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestSeeders_Postgres(t *testing.T) {
	// Integration tests usually shouldn't run during 'go test ./...'
	// unless specified, but we'll run it here for your coverage.
	db := setupPostgresDB(t)

	t.Run("Full Success Path", func(t *testing.T) {
		// This triggers SeedUsers and SeedTask
		Run(db)

		var userCount int64
		var taskCount int64
		db.Model(&models.User{}).Count(&userCount)
		db.Model(&models.Task{}).Count(&taskCount)

		// Your seeder adds 2 users and 50 tasks
		assert.Equal(t, int64(2), userCount)
		assert.Equal(t, int64(50), taskCount)
	})

	t.Run("SeedTask - No Users Found", func(t *testing.T) {
		// Clean the database
		db.Exec("DELETE FROM tasks")
		db.Exec("DELETE FROM users")

		// Should hit the log.Println("❌ Skipping task seeding...")
		SeedTask(db)

		var taskCount int64
		db.Model(&models.Task{}).Count(&taskCount)
		assert.Equal(t, int64(0), taskCount)
	})
}

func TestSeeder_PostgresFailures(t *testing.T) {
	db := setupPostgresDB(t)

	t.Run("SeedUsers - DB Error", func(t *testing.T) {
		// Drop table to force the db.Create(&users).Error block
		db.Migrator().DropTable(&models.User{})

		SeedUsers(db)

		assert.False(t, db.Migrator().HasTable(&models.User{}))

		// Re-migrate for the next test
		db.AutoMigrate(&models.User{})
	})

	t.Run("SeedTask - DB Error", func(t *testing.T) {
		// Ensure users exist so we get past the first check
		db.AutoMigrate(&models.User{})
		db.Create(&models.User{Name: "Test User", Email: "test@ex.com"})

		// Drop task table to force the result.Error != nil block
		db.Migrator().DropTable(&models.Task{})

		SeedTask(db)

		assert.False(t, db.Migrator().HasTable(&models.Task{}))

		// Final cleanup: Restore tables
		db.AutoMigrate(&models.Task{})
	})
}
