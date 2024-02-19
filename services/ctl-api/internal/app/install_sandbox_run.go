package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type SandboxRunType string

const (
	SandboxRunTypeProvision   SandboxRunType = "provision"
	SandboxRunTypeReprovision SandboxRunType = "reprovision"
	SandboxRunTypeDeprovision SandboxRunType = "deprovision"
)

type InstallSandboxRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// used for RLS
	OrgID     string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org       Org    `json:"-" faker:"-"`
	InstallID string `json:"install_id" gorm:"not null;default null"`

	RunType           SandboxRunType `json:"run_type"`
	Status            string         `json:"status" gorm:"notnull"`
	StatusDescription string         `json:"status_description" gorm:"notnull"`

	AppSandboxConfigID string           `json:"-"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config"`
}

func (i *InstallSandboxRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.CreatedByID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
