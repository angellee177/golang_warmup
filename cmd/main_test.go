package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain handles the global state for the 'cmd' package tests.
// This ensures JWT_SECRET is present before any routes are initialized.
func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "required-test-secret-key")
	os.Setenv("GIN_MODE", "release")
	os.Setenv("PORT", "8080")

	exitCode := m.Run()
	os.Exit(exitCode)
}

// TestInitialize is now fast because it uses t.Setenv for local overrides
func TestInitialize(t *testing.T) {
	tests := []struct {
		name         string
		portEnv      string
		expectedPort string
	}{
		{"Default port when empty", "", "8080"},
		{"Custom port from env", "9090", "9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.portEnv)

			// We pass 'true' to skip DB migrations, but Initialize still calls config.Init()
			// If config.Init() is slow, this test will be slow.
			app := Initialize(true)

			assert.NotNil(t, app.Router)
			assert.Equal(t, tt.expectedPort, app.Port)
		})
	}
}

// TestAPIEndpoints is lightning fast because it bypasses Initialize() and DB entirely
func TestAPIEndpoints(t *testing.T) {
	// We pass nil to SetupRouter. Ensure SetupRoutes handles a nil DB!
	router := SetupRouter(nil)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{"Health Check Success", "/health", http.StatusOK, "Ping!"},
		{"Route Not Found", "/undefined-route", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// TestMainLogic is your only "Slow" test because it connects to Postgres
func TestMainLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := Initialize(false) // Triggers real migrations/seeds

	assert.NotNil(t, app.DB)
	assert.Equal(t, "8080", app.Port)
}
