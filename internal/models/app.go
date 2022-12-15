package models

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Model

	CreatedByID     uuid.UUID `gorm:"type:uuid"`
	Name            string
	Slug            string
	OrgID           uuid.UUID
	GithubInstallID string

	Components []Component `fake:"skip"`
	Installs   []Install   `fake:"skip"`
}

func (App) IsNode() {}

func (a App) GetID() string {
	return a.Model.ID.String()
}

func (a App) GetCreatedAt() time.Time {
	return a.Model.CreatedAt
}

func (a App) GetUpdatedAt() time.Time {
	return a.Model.UpdatedAt
}
