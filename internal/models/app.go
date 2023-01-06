package models

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Model

	CreatedByID     string
	Name            string
	Slug            string
	OrgID           uuid.UUID
	GithubInstallID string

	Components []Component `faker:"-"`
	Installs   []Install   `faker:"-"`
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
