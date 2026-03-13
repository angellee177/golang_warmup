package service

import (
	"context"
	"errors"
	"testing"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockTaskRepo struct {
	mock.Mock
}

// Satisfying ITaskRepository (and Generic)
func (m *MockTaskRepo) Create(ctx context.Context, entity *models.Task) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTaskRepo) FindByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*models.Task, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepo) SearchTasks(ctx context.Context, t, a string, s models.TaskStatus, search string, l, o int, uid uuid.UUID) ([]models.Task, int64, error) {
	args := m.Called(ctx, t, a, s, search, l, o, uid)
	return args.Get(0).([]models.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockTaskRepo) Update(ctx context.Context, entity *models.Task) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mocking required generic methods unused by service but needed for interface
func (m *MockTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return nil, nil
}
func (m *MockTaskRepo) FindAll(ctx context.Context, l, o int) ([]models.Task, int64, error) {
	return nil, 0, nil
}

func TestTaskService(t *testing.T) {
	repo := new(MockTaskRepo)
	svc := NewTaskService(repo)
	ctx := context.Background()
	uID := uuid.New()
	tID := uuid.New()

	t.Run("CreateTask - Default Status", func(t *testing.T) {
		task := &models.Task{Title: "New Task"}
		repo.On("Create", ctx, task).Return(nil).Once()

		err := svc.CreateTask(ctx, task, uID)

		assert.NoError(t, err)
		assert.Equal(t, models.StatusTodo, task.Status)
		assert.Equal(t, uID, task.UserID)
	})

	t.Run("GetTaskByID - Success", func(t *testing.T) {
		repo.On("FindByIDAndUser", ctx, tID, uID).Return(&models.Task{Title: "Found"}, nil).Once()

		res, err := svc.GetTaskByID(ctx, tID, uID)
		assert.NoError(t, err)
		assert.Equal(t, "Found", res.Title)
	})

	t.Run("GetTaskByID - Not Found", func(t *testing.T) {
		repo.On("FindByIDAndUser", ctx, tID, uID).Return(nil, errors.New("db error")).Once()

		res, err := svc.GetTaskByID(ctx, tID, uID)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "task not found or access denied", err.Error())
	})

	t.Run("UpdateTask - Success", func(t *testing.T) {
		existing := &models.Task{ID: tID, Title: "Old", UserID: uID}
		input := &models.Task{Title: "New"}

		repo.On("FindByIDAndUser", ctx, tID, uID).Return(existing, nil).Once()
		repo.On("Update", ctx, existing).Return(nil).Once()

		res, err := svc.UpdateTask(ctx, tID, input, uID)
		assert.NoError(t, err)
		assert.Equal(t, "New", res.Title)
	})

	t.Run("UpdateTask - Fail because Task Not Found", func(t *testing.T) {
		input := &models.Task{Title: "Updated Title"}
		repo.On("FindByIDAndUser", ctx, tID, uID).
			Return(nil, errors.New("not found")).Once()

		res, err := svc.UpdateTask(ctx, tID, input, uID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "task not found or access denied", err.Error())
	})

	t.Run("DeleteTask - Fail because Task Not Found", func(t *testing.T) {
		repo.On("FindByIDAndUser", ctx, tID, uID).
			Return(nil, errors.New("not found")).Once()

		err := svc.DeleteTask(ctx, tID, uID)

		assert.Error(t, err)
		assert.Equal(t, "task not found or access denied", err.Error())

		repo.AssertNotCalled(t, "Delete", ctx, tID)
	})

	t.Run("DeleteTask - Success", func(t *testing.T) {
		repo.On("FindByIDAndUser", ctx, tID, uID).Return(&models.Task{}, nil).Once()
		repo.On("Delete", ctx, tID).Return(nil).Once()

		err := svc.DeleteTask(ctx, tID, uID)
		assert.NoError(t, err)
	})

	t.Run("GetTasks - Success", func(t *testing.T) {
		repo.On("SearchTasks", ctx, "t", "a", models.StatusTodo, "s", 10, 0, uID).
			Return([]models.Task{}, int64(0), nil).Once()

		_, _, err := svc.GetTasks(ctx, "t", "a", models.StatusTodo, "s", 10, 0, uID)
		assert.NoError(t, err)
	})
}
