package service

import (
	"context"
	"errors"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/angellee177/go-tasks-crud/repository"
	"github.com/google/uuid"
)

type ITaskService interface {
	CreateTask(ctx context.Context, task *models.Task, currentUserID uuid.UUID) error
	GetTaskByID(
		ctx context.Context,
		id uuid.UUID,
		currentUserID uuid.UUID,
	) (*models.Task, error)
	GetTasks(
		ctx context.Context,
		title string,
		author string,
		status models.TaskStatus,
		search string,
		limit int,
		offset int,
		currentUserID uuid.UUID,
	) ([]models.Task, int64, error)
	UpdateTask(ctx context.Context, id uuid.UUID, input *models.Task, currentUserID uuid.UUID) (*models.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) error
}

type taskService struct {
	repo repository.ITaskRepository
}

func NewTaskService(repo repository.ITaskRepository) ITaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, task *models.Task, currentUserID uuid.UUID) error {
	task.UserID = currentUserID

	// set default status
	if task.Status == "" {
		task.Status = models.StatusTodo
	}

	return s.repo.Create(ctx, task)
}

func (s *taskService) GetTaskByID(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) (*models.Task, error) {
	task, err := s.repo.FindByIDAndUser(ctx, id, currentUserID)

	if err != nil {
		return nil, errors.New("task not found or access denied")
	}
	return task, nil
}

func (s *taskService) GetTasks(
	ctx context.Context,
	title string,
	author string,
	status models.TaskStatus,
	search string,
	limit int,
	offset int,
	currentUserID uuid.UUID,
) ([]models.Task, int64, error) {
	return s.repo.SearchTasks(
		ctx,
		title,
		author,
		status,
		search,
		limit,
		offset,
		currentUserID,
	)
}

func (s *taskService) UpdateTask(ctx context.Context, id uuid.UUID, input *models.Task, currentUserID uuid.UUID) (*models.Task, error) {
	// Reuse the FindByIDAndUser logic
	task, err := s.GetTaskByID(ctx, id, currentUserID)
	if err != nil {
		return nil, err
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Status = input.Status
	task.DueDate = input.DueDate

	err = s.repo.Update(ctx, task)
	return task, err
}

func (s *taskService) DeleteTask(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) error {
	// Reuse the FindByIDAndUser logic
	_, err := s.GetTaskByID(ctx, id, currentUserID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
