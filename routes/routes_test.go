package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
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
