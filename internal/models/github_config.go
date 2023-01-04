package models

import (
	"time"

	"github.com/google/uuid"
)

type GithubConfig struct {
	Model

	ComponentID uuid.UUID

	Repo      string `json:"repo"`
	Directory string `json:"directory"`
	RepoOwner string `json:"repo_owner"`
	Branch    string `json:"branch"`
}

func (GithubConfig) IsVcsConfig() {}

func (GithubConfig) IsNode() {}

func (gitCfg GithubConfig) GetID() string {
	return gitCfg.Model.ID.String()
}

func (gitCfg GithubConfig) GetCreatedAt() time.Time {
	return gitCfg.Model.CreatedAt
}

func (gitCfg GithubConfig) GetUpdatedAt() time.Time {
	return gitCfg.Model.UpdatedAt
}
