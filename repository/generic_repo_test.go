package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gorm_pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestEntity matches the interface requirements (needs a UUID and created_at for FindAll order)
type TestEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"unique"`
	CreatedAt time.Time
}

func setupGenericTestContainer(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	// 1. Start Postgres Container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Connect GORM
	connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	db, _ := gorm.Open(gorm_pg.Open(connStr), &gorm.Config{})

	// Migrate our test entity
	db.AutoMigrate(&TestEntity{})

	return db, func() { pgContainer.Terminate(ctx) }
}

func TestGenericRepository_Implementation(t *testing.T) {
	db, cleanup := setupGenericTestContainer(t)
	defer cleanup()

	repo := NewGenericRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("Create and FindByID", func(t *testing.T) {
		id := uuid.New()
		entity := &TestEntity{ID: id, Name: "Generic Test"}

		err := repo.Create(ctx, entity)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "Generic Test", found.Name)
	})

	t.Run("FindAll with Pagination", func(t *testing.T) {
		// Insert a few items
		for i := 0; i < 5; i++ {
			repo.Create(ctx, &TestEntity{ID: uuid.New(), Name: string(rune(i))})
		}

		entities, total, err := repo.FindAll(ctx, 2, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total) // 1 from prev test + 5 here
		assert.Len(t, entities, 2)       // Limit was 2
	})

	t.Run("Update", func(t *testing.T) {
		id := uuid.New()
		entity := &TestEntity{ID: id, Name: "Old Name"}
		repo.Create(ctx, entity)

		entity.Name = "New Name"
		err := repo.Update(ctx, entity)
		assert.NoError(t, err)

		found, _ := repo.FindByID(ctx, id)
		assert.Equal(t, "New Name", found.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		id := uuid.New()
		repo.Create(ctx, &TestEntity{ID: id, Name: "Delete Me"})

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Error Path - Duplicate Unique Key", func(t *testing.T) {
		// This covers the error return path in Create
		name := "Duplicate"
		repo.Create(ctx, &TestEntity{ID: uuid.New(), Name: name})

		err := repo.Create(ctx, &TestEntity{ID: uuid.New(), Name: name})
		assert.Error(t, err) // Postgres Unique Violation
	})
}
