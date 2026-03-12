package models

import (
	"errors"
	"time"

	"github.com/angellee177/go-tasks-crud/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type Task struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `gorm:"type:text;default:todo" json:"status"`
	// Use the custom Date type here
	DueDate common.Date `gorm:"type:date;not null" json:"due_date" binding:"required"`

	UserID uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"creator,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Task) BeforeSave(tx *gorm.DB) (err error) {
	dueTime := time.Time(t.DueDate) // Cast custom Date back to time.Time

	if dueTime.IsZero() {
		return errors.New("due date is required")
	}

	// Comparison logic
	if dueTime.Before(time.Now().Truncate(24 * time.Hour)) {
		return errors.New("due date cannot be in the past")
	}
	return nil
}
