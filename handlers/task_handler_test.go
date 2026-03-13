package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, task *models.Task, userID uuid.UUID) error {
	// IMPORTANT: Use ctx (context.Context), not a string
	args := m.Called(ctx, task, userID)
	return args.Error(0)
}

func (m *MockTaskService) GetTasks(ctx context.Context, title, author string, status models.TaskStatus, search string, limit, offset int, userID uuid.UUID) ([]models.Task, int64, error) {
	args := m.Called(ctx, title, author, status, search, limit, offset, userID)
	return args.Get(0).([]models.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockTaskService) GetTaskByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, id uuid.UUID, task *models.Task, userID uuid.UUID) (*models.Task, error) {
	args := m.Called(ctx, id, task, userID)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func TestTaskHandler_AddTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success with Custom Date", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		userID := uuid.New()

		// Setup the payload exactly how a frontend would send it
		// Use a future date to satisfy the BeforeSave hook logic
		futureDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

		inputJSON := map[string]interface{}{
			"title":    "Finish Unit Tests",
			"due_date": futureDate, // "2026-03-20"
		}

		mockService.On("CreateTask",
			mock.Anything,
			mock.AnythingOfType("*models.Task"),
			userID,
		).Return(nil)

		w := httptest.NewRecorder()
		r := gin.New()

		r.POST("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.AddTask(c)
		})

		body, _ := json.Marshal(inputJSON)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Fail - Past Due Date", func(t *testing.T) {
		// Note: The BeforeSave hook logic is in the model, but usually
		// the service layer calls that. To test the handler's reaction to
		// a service error (like "due date cannot be in the past"):

		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)
		userID := uuid.New()

		mockService.On("CreateTask", mock.Anything, mock.Anything, userID).
			Return(assert.AnError) // Simulate service returning hook error

		inputJSON := map[string]interface{}{
			"title":    "Past Task",
			"due_date": "2020-01-01",
		}

		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.AddTask(c)
		})

		body, _ := json.Marshal(inputJSON)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Handler returns 500 "Could not create task" on service error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Binding Error - Missing Required Field DueDate", func(t *testing.T) {
		// We don't need a mock service here because the code returns
		// BEFORE reaching the service layer.
		handler := NewTaskHandler(nil)

		w := httptest.NewRecorder()
		r := gin.New()

		r.POST("/tasks", handler.AddTask)

		// Create a payload missing 'due_date' (which has binding:"required")
		payload := map[string]interface{}{
			"title": "Missing Date Task",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Field validation for 'DueDate' failed")
	})

	t.Run("Binding Error - Invalid JSON Syntax", func(t *testing.T) {
		handler := NewTaskHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()
		r.POST("/tasks", handler.AddTask)

		// Sending completely broken JSON syntax
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBufferString("{ invalid: json }"))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTaskHandler_GetTaskDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	taskID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		expectedTask := &models.Task{Title: "Detail Task"}
		mockService.On("GetTaskByID", mock.Anything, taskID, userID).Return(expectedTask, nil)

		w := httptest.NewRecorder()
		r := gin.New()

		r.GET("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTaskDetail(c)
		})

		req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Detail Task")
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID Format", func(t *testing.T) {
		handler := NewTaskHandler(nil) // Service shouldn't even be called
		w := httptest.NewRecorder()
		r := gin.New()

		r.GET("/tasks/:id", handler.GetTaskDetail)

		// Pass a non-UUID string
		req, _ := http.NewRequest("GET", "/tasks/not-a-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid task ID format")
	})

	t.Run("Unauthorized - UserID Missing in Context", func(t *testing.T) {
		handler := NewTaskHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()

		r.GET("/tasks/:id", func(c *gin.Context) {
			// We EXPLICITLY do NOT set "userId" here
			handler.GetTaskDetail(c)
		})

		req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("Task Not Found in Service", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		// Mock returns an error to simulate DB not finding the record
		mockService.On("GetTaskByID", mock.Anything, taskID, userID).
			Return((*models.Task)(nil), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()

		r.GET("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTaskDetail(c)
		})

		req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Task not found")
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	taskID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		futureDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		inputJSON := map[string]interface{}{
			"title":    "Updated Task Title",
			"due_date": futureDate,
		}

		expectedTask := &models.Task{
			ID:    taskID,
			Title: "Updated Task Title",
		}

		mockService.On("UpdateTask", mock.Anything, taskID, mock.AnythingOfType("*models.Task"), userID).
			Return(expectedTask, nil)

		w := httptest.NewRecorder()
		r := gin.New()

		r.PUT("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.UpdateTask(c)
		})

		body, _ := json.Marshal(inputJSON)
		req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Updated Task Title")
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		handler := NewTaskHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()

		r.PUT("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", uuid.New().String())
			handler.UpdateTask(c)
		})

		req, _ := http.NewRequest("PUT", "/tasks/not-a-uuid", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid UUID")
	})

	t.Run("Binding Error - Missing DueDate", func(t *testing.T) {
		handler := NewTaskHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()

		r.PUT("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.UpdateTask(c)
		})

		// Missing "due_date" triggers 400 because of binding:"required"
		payload := map[string]interface{}{"title": "New Title"}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		mockService.On("UpdateTask", mock.Anything, taskID, mock.Anything, userID).
			Return((*models.Task)(nil), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()

		r.PUT("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.UpdateTask(c)
		})

		futureDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		body, _ := json.Marshal(map[string]interface{}{"title": "Title", "due_date": futureDate})
		req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	taskID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		mockService.On("DeleteTask", mock.Anything, taskID, userID).Return(nil)

		w := httptest.NewRecorder()
		r := gin.New()

		r.DELETE("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.DeleteTask(c)
		})

		req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task Deleted successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		handler := NewTaskHandler(nil)
		w := httptest.NewRecorder()
		r := gin.New()

		r.DELETE("/tasks/:id", handler.DeleteTask)

		req, _ := http.NewRequest("DELETE", "/tasks/not-an-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid UUID")
	})

	t.Run("Service Failure", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		// Mock the service returning an error (e.g., DB failure or unauthorized access)
		mockService.On("DeleteTask", mock.Anything, taskID, userID).
			Return(assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()

		r.DELETE("/tasks/:id", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.DeleteTask(c)
		})

		req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_GetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	t.Run("Success with Pagination and Filters", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		// Setup expectations
		expectedTasks := []models.Task{
			{Title: "Task 1", Status: models.StatusTodo},
			{Title: "Task 2", Status: models.StatusTodo},
		}
		var totalResults int64 = 25 // 25 total items means 3 pages at limit 10

		// Mock expectation matching the handler's default math
		// page 2, limit 10 -> offset 10
		mockService.On("GetTasks",
			mock.Anything,             // ctx
			"TestTitle",               // title
			"AuthorName",              // author
			models.TaskStatus("todo"), // status
			"search-term",             // search
			10,                        // limit
			10,                        // offset
			userID,                    // userID
		).Return(expectedTasks, totalResults, nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTasks(c)
		})

		// Construct URL with query params
		req, _ := http.NewRequest("GET", "/tasks?page=2&limit=10&title=TestTitle&author=AuthorName&status=todo&search=search-term", nil)
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, float64(25), response["total_results"])
		assert.Equal(t, float64(3), response["total_pages"]) // Ceil(25/10) = 3
		assert.Equal(t, float64(2), response["page"])
		mockService.AssertExpectations(t)
	})

	t.Run("Default Pagination Values", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		// Expect defaults: page=1 (offset 0), limit=10
		mockService.On("GetTasks", mock.Anything, "", "", models.TaskStatus(""), "", 10, 0, userID).
			Return([]models.Task{}, int64(0), nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTasks(c)
		})

		req, _ := http.NewRequest("GET", "/tasks", nil) // No params
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative Page Correction", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		// Handler should convert page -5 to page 1 (offset 0)
		mockService.On("GetTasks", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 10, 0, userID).
			Return([]models.Task{}, int64(0), nil)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTasks(c)
		})

		req, _ := http.NewRequest("GET", "/tasks?page=-5", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService := new(MockTaskService)
		handler := NewTaskHandler(mockService)

		mockService.On("GetTasks", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, userID).
			Return([]models.Task{}, int64(0), assert.AnError)

		w := httptest.NewRecorder()
		r := gin.New()
		r.GET("/tasks", func(c *gin.Context) {
			c.Set("userId", userID.String())
			handler.GetTasks(c)
		})

		req, _ := http.NewRequest("GET", "/tasks", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), assert.AnError.Error())
	})
}
