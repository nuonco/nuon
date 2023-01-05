// component.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Component struct {
	Model

	Name  string
	AppID uuid.UUID
	App   App `faker:"-"`

	BuildImage      string `json:"container_image_url"`
	Type            string `json:"type"`
	GithubRepo      string `json:"github_repo"`
	GithubDir       string `json:"github_dir"`
	GithubRepoOwner string `json:"github_repo_owner"`
	GithubBranch    string `json:"github_branch"`

	Deployments  []Deployment  `faker:"-"`
	VcsConfig    VcsConfig     `gorm:"-" faker:"-"`
	GithubConfig *GithubConfig `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" faker:"-"`
}

func (Component) IsNode() {}

func (c Component) GetID() string {
	return c.Model.ID.String()
}

func (c Component) GetCreatedAt() time.Time {
	return c.Model.CreatedAt
}

func (c Component) GetUpdatedAt() time.Time {
	return c.Model.UpdatedAt
}
