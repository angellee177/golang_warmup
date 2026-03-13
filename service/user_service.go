package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/angellee177/go-tasks-crud/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type IUserService interface {
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, email, password string) (string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error)
}

type userService struct {
	repo repository.IGenericRepository[models.User]
}

func NewUserService(repo repository.IGenericRepository[models.User]) IUserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, user *models.User) error {
	if user.Password == "" {
		return errors.New("password cannot be empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashed)
	return s.repo.Create(ctx, user)
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// Specialized lookup using Type Assertion (Same as Task)
	user, err := s.repo.(repository.IUserRepository).FindByEmail(ctx, email)

	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	return s.repo.FindAll(ctx, limit, offset)
}
