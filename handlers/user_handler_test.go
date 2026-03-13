package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService defines the mock for user operations
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestUserHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("Register", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
			return u.Email == "test@example.com"
		})).Return(nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/register", handler.Register)

		payload := map[string]string{"email": "test@example.com", "password": "password123"}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Binding Error", func(t *testing.T) {
		handler := NewUserHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/register", handler.Register)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString("{invalid-json}"))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Service Failure", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		// Setup the mock to return an error when Register is called
		mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "fail@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), assert.AnError.Error())
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("Login", mock.Anything, "test@example.com", "password123").
			Return("mock-jwt-token", nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/login", handler.Login)

		payload := map[string]string{"email": "test@example.com", "password": "password123"}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "mock-jwt-token")
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).
			Return("", assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/login", handler.Login)

		payload := map[string]string{"email": "wrong@ex.com", "password": "wrong"}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Binding Error - Missing Fields", func(t *testing.T) {
		handler := NewUserHandler(nil) // Service not needed for binding errors
		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/login", handler.Login)

		// Scenario 1: Missing 'password' which is required in struct
		payload := map[string]string{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Field validation for 'Password' failed")
	})

	t.Run("Binding Error - Invalid JSON Syntax", func(t *testing.T) {
		handler := NewUserHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/login", handler.Login)

		// Scenario 2: Broken JSON syntax
		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("{ invalid json }"))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_GetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("GetAllUsers", mock.Anything, 10, 0).
			Return([]models.User{{Email: "user1@ex.com"}}, int64(1), nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/users", handler.GetUsers)

		req, _ := http.NewRequest("GET", "/users?limit=10&page=1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user1@ex.com")
	})

	t.Run("Service Failure", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		// Mock expectation: When GetAllUsers is called with any arguments, return an error
		mockService.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything).
			Return([]models.User{}, int64(0), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/users", handler.GetUsers)

		// Make the request
		req, _ := http.NewRequest("GET", "/users?limit=10&page=1", nil)
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), assert.AnError.Error())
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("GetUserByID", mock.Anything, userID).
			Return(&models.User{Email: "profile@ex.com"}, nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/profile", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetProfile(c)
		})

		req, _ := http.NewRequest("GET", "/profile", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "profile@ex.com")
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		userID := uuid.New()

		// Mock the service to return nil for the user and an error
		mockService.On("GetUserByID", mock.Anything, userID).
			Return((*models.User)(nil), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/profile", func(c *gin.Context) {
			// Simulate the middleware providing the userID
			c.Set("userId", userID.String())
			handler.GetProfile(c)
		})

		req, _ := http.NewRequest("GET", "/profile", nil)
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserById(t *testing.T) {
	gin.SetMode(gin.TestMode)
	targetID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)

		mockService.On("GetUserByID", mock.Anything, targetID).
			Return(&models.User{Email: "findme@ex.com"}, nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/users/:id", handler.GetUserById)

		req, _ := http.NewRequest("GET", "/users/"+targetID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		handler := NewUserHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/users/:id", handler.GetUserById)

		req, _ := http.NewRequest("GET", "/users/not-a-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User Not Found in Database", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := NewUserHandler(mockService)
		targetID := uuid.New()

		// Mock: Valid UUID but Service returns an error (e.g., record not found)
		mockService.On("GetUserByID", mock.Anything, targetID).
			Return((*models.User)(nil), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/users/:id", handler.GetUserById)

		// Use a valid UUID string so it passes the first uuid.Parse check
		req, _ := http.NewRequest("GET", "/users/"+targetID.String(), nil)
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
		mockService.AssertExpectations(t)
	})
}
