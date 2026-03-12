package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IGenericRepository[T any] interface {
	// The interface says it needs context
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	FindAll(ctx context.Context, limit, offset int) ([]T, int64, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repoImpl[T any] struct {
	db *gorm.DB
}

func NewGenericRepository[T any](db *gorm.DB) IGenericRepository[T] {
	return &repoImpl[T]{db: db}
}

func (r *repoImpl[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *repoImpl[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	return &entity, err
}

func (r *repoImpl[T]) FindAll(ctx context.Context, limit, offset int) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.db.WithContext(ctx).Model(new(T))
	db.Count(&total)

	err := db.Limit(limit).Offset(offset).Order("created_at desc").Find(&entities).Error
	return entities, total, err
}

func (r *repoImpl[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *repoImpl[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error
}
