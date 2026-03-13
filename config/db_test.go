package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInit_Success covers the successful connection path
func TestInit_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := Init()
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestInit_EnvFailure(t *testing.T) {
	_, b, _, _ := runtime.Caller(0)
	envPath := filepath.Join(filepath.Dir(b), "../.env")

	// 1. Sabotage permissions
	info, err := os.Stat(envPath)
	if err == nil {
		originalMode := info.Mode()
		os.Chmod(envPath, 0000)
		defer os.Chmod(envPath, originalMode)
	}

	if os.Getenv("BE_CRASHER") == "1" {
		_, _ = Init()
		return
	}

	// 2. Run sub-process and CAPTURE the error
	cmd := exec.Command(os.Args[0], "-test.run=TestInit_EnvFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	// THIS IS THE KEY CHANGE: assign the result to err
	err = cmd.Run()

	// 3. Verify the crash
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return // SUCCESS: The sub-process crashed as expected
	}

	t.Fatalf("Expected Init to fail, but it didn't.")
}

func TestInit_ConnectionFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "2" {
		// Sabotage environment variables so GORM fails to connect
		os.Setenv("DB_PORT", "1") // Impossible port
		_, _ = Init()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInit_ConnectionFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=2")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("Expected Init to fail due to invalid DB port")
}
