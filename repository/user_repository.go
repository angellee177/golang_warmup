package repository

import (
	"context"

	"github.com/angellee177/go-tasks-crud/models"
	"gorm.io/gorm"
)

type IUserRepository interface {
	IGenericRepository[models.User]
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepoImpl struct {
	IGenericRepository[models.User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &userRepoImpl{
		IGenericRepository: NewGenericRepository[models.User](db),
		db:                 db,
	}
}

func (r *userRepoImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}
