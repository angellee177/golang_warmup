package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/angellee177/go-tasks-crud/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock IUserRepository
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) FindAll(ctx context.Context, l, o int) ([]models.User, int64, error) {
	args := m.Called(ctx, l, o)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// These satisfy the interface but aren't used in UserService
func (m *MockUserRepo) Update(ctx context.Context, u *models.User) error { return nil }
func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error   { return nil }

func TestUserService(t *testing.T) {
	repo := new(MockUserRepo)
	svc := NewUserService(repo)
	ctx := context.Background()
	os.Setenv("JWT_SECRET", "test_secret")

	t.Run("Register - Success", func(t *testing.T) {
		user := &models.User{Password: "secure123", Email: "test@test.com"}
		repo.On("Create", ctx, user).Return(nil).Once()

		err := svc.Register(ctx, user)

		assert.NoError(t, err)
		assert.NotEqual(t, "secure123", user.Password) // Verify hashing happened
		repo.AssertExpectations(t)
	})

	t.Run("Register - Empty Password Error", func(t *testing.T) {
		user := &models.User{Password: ""}
		err := svc.Register(ctx, user)
		assert.Error(t, err)
		assert.Equal(t, "password cannot be empty", err.Error())
	})

	t.Run("Register - Hashing Failure (Password too long)", func(t *testing.T) {
		// bcrypt has a maximum password length of 72 bytes.
		// Providing a significantly larger string will cause GenerateFromPassword to error.
		longPassword := make([]byte, 100)
		for i := range longPassword {
			longPassword[i] = 'a'
		}

		user := &models.User{
			Password: string(longPassword),
			Email:    "too-long@test.com",
		}

		// Act
		err := svc.Register(ctx, user)

		// Assert: Hits 'if err != nil { return fmt.Errorf(...) }'
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to hash password")
	})

	t.Run("Login - Success", func(t *testing.T) {
		pass := "password123"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		dbUser := &models.User{ID: uuid.New(), Email: "test@test.com", Password: string(hashed)}

		repo.On("FindByEmail", ctx, "test@test.com").Return(dbUser, nil).Once()

		token, err := svc.Login(ctx, "test@test.com", pass)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Login - Invalid Email", func(t *testing.T) {
		repo.On("FindByEmail", ctx, "wrong@test.com").Return(nil, errors.New("not found")).Once()

		token, err := svc.Login(ctx, "wrong@test.com", "any")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("Login - Wrong Password", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		dbUser := &models.User{Email: "test@test.com", Password: string(hashed)}

		repo.On("FindByEmail", ctx, "test@test.com").Return(dbUser, nil).Once()

		token, err := svc.Login(ctx, "test@test.com", "wrong_pass")

		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
		assert.Empty(t, token)
	})

	t.Run("GetUserByID - Success", func(t *testing.T) {
		id := uuid.New()
		repo.On("FindByID", ctx, id).Return(&models.User{ID: id}, nil).Once()

		res, err := svc.GetUserByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, id, res.ID)
	})

	t.Run("GetAllUsers - Success", func(t *testing.T) {
		repo.On("FindAll", ctx, 10, 0).Return([]models.User{}, int64(0), nil).Once()

		users, total, err := svc.GetAllUsers(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, users, 0)
	})
}
