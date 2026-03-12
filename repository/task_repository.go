package repository

import (
	"context"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ITaskRepository interface {
	IGenericRepository[models.Task] // Inherit basic CRUD
	FindByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	SearchTasks(
		ctx context.Context,
		title string,
		author string,
		status models.TaskStatus,
		search string,
		limit int,
		offset int,
		currentUserID uuid.UUID,
	) ([]models.Task, int64, error)
}

type taskRepoImpl struct {
	IGenericRepository[models.Task] // Composition
	db                              *gorm.DB
}

func NewTaskRepository(db *gorm.DB) ITaskRepository {
	return &taskRepoImpl{
		IGenericRepository: NewGenericRepository[models.Task](db),
		db:                 db,
	}
}

func (r *taskRepoImpl) FindByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task

	// If the combination of Task ID and User ID doesn't exist, GORM returns ErrRecordNotFound
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ? AND user_id = ?", id, userID).
		First(&task).Error

	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepoImpl) SearchTasks(
	ctx context.Context,
	title string,
	author string,
	status models.TaskStatus,
	search string,
	limit int,
	offset int,
	currentUserID uuid.UUID,
) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// user Preload("User") to fetch the related user data
	query := r.db.WithContext(ctx).Model(&models.Task{}).
		Preload("User").
		Where("user_id = ?", currentUserID)

	if author != "" {
		query = query.Joins("JOIN users ON users.id = tasks.user_id").
			Where("users.name ILIKE ?", "%"+author+"%")
	}
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if status != "" {
		query = query.Where("status ILIKE ?", "%"+status+"%")
	}
	if search != "" {
		s := "%" + search + "%"
		query = query.Where(r.db.Where("tasks.title ILIKE ?", s).Or("tasks.description ILIKE ?", s))
	}

	query.Count(&total)
	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&tasks).Error

	return tasks, total, err
}
