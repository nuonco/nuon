package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) GetJobsManager() (jobs.Manager, error) {
	// TODO(jm): update this
	return nil, nil
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	return
}

type IDer interface {
	GetID() string
}

func (m *Model) GetID() string {
	return m.ID.String()
}
