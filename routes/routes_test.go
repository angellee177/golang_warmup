package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMain ensures the environment is ready for the routes package
func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("GIN_MODE", "test")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use nil for DB to keep it as a pure unit test
	SetupRoutes(router, nil)

	routes := router.Routes()

	expectedRoutes := []struct {
		Method string
		Path   string
	}{
		{"POST", "/v1/tasks/"},
		{"GET", "/v1/tasks/"},
		{"POST", "/v1/users/register"},
		{"POST", "/v1/users/login"},
	}

	for _, expected := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Path == expected.Path && route.Method == expected.Method {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s %s should be registered", expected.Method, expected.Path)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/undefined", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserRoutes_PanicsOnMissingSecret(t *testing.T) {
	// check if an environment variable is set to avoid infinite recursion
	if os.Getenv("BE_CRASHER") == "1" {
		os.Unsetenv("JWT_SECRET") // Ensure it's empty

		router := gin.New()
		v1 := router.Group("/v1")

		// This should trigger log.Fatal
		UserRoutes(v1, nil)
		return
	}

	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestUserRoutes_PanicsOnMissingSecret")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	err := cmd.Run()

	// Assert that the process exited with an error (log.Fatal causes exit status 1)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return // Successfully triggered the Fatal branch!
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
