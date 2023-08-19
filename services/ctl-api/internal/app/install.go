package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Install struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name   string `json:"name"`
	App    App    `swaggerignore:"true" json:"-"`
	AppID  string `json:"app_id"`
	Status string `json:"status"`

	SandboxReleaseID string         `json:"-"`
	SandboxRelease   SandboxRelease `json:"sandbox_release"`

	InstallComponents []InstallComponent `json:"install_components,omitempty"`
	AWSAccount        AWSAccount         `json:"aws_account"`
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	i.ID = domains.NewInstallID()
	return nil
}
