package migrations

import (
	"os"
	"os/exec"
	"testing"

	"github.com/angellee177/go-tasks-crud/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRunMigrations_Success(t *testing.T) {
	// Skip if we are in a CI environment without a real Postgres instance
	if testing.Short() {
		t.Skip("Skipping migration test in short mode")
	}

	// 1. Initialize the real DB connection from config
	db, err := config.Init()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 2. Run the migration
	// This covers the log.Println, db.Exec, and AutoMigrate lines
	RunMigrations(db)

	// 3. Verify tables exist
	assert.True(t, db.Migrator().HasTable("tasks"))
	assert.True(t, db.Migrator().HasTable("users"))
}

func TestRunMigrations_FatalError(t *testing.T) {
	// Special pattern to cover the log.Fatalf path
	if os.Getenv("BE_CRASHER") == "1" {
		// Pass a nil or closed DB to force an error in AutoMigrate
		RunMigrations(nil)
		return
	}

	// Run the test in a separate process to catch the Fatal exit
	cmd := exec.Command(os.Args[0], "-test.run=TestRunMigrations_FatalError")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	// We expect an error (exit status 1) because log.Fatalf was called
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Process should have crashed with log.Fatalf, but it didn't")
}

func TestRunMigrations_Failure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		// 1. We create a real GORM DB instance but use a "dummy" dialector
		// or a DSN that is syntactically valid but unreachable.
		// Or better: Use a DB that is already closed.

		dialector := postgres.Open("host=localhost port=1 user=invalid password=invalid dbname=invalid sslmode=disable")
		db, _ := gorm.Open(dialector, &gorm.Config{})

		// Force an error by closing the underlying SQL connection immediately
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}

		// 2. Now RunMigrations will get an error from db.Exec or db.AutoMigrate
		RunMigrations(db)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunMigrations_Failure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	// We expect exit status 1
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Process should have hit log.Fatalf and exited with 1")
}
