package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Install struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	Name              string `json:"name" gorm:"notnull;index:idx_app_install_name,unique"`
	App               App    `swaggerignore:"true" json:"app"`
	AppID             string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique"`
	Status            string `json:"status" gorm:"notnull"`
	StatusDescription string `json:"status_description" gorm:"notnull"`

	AppSandboxConfigID string           `json:"-" swaggerignore:"true"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config"`

	AppRunnerConfigID string          `json:"-" swaggerignore:"true"`
	AppRunnerConfig   AppRunnerConfig `json:"app_runner_config"`

	InstallComponents  []InstallComponent  `json:"install_components,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	InstallSandboxRuns []InstallSandboxRun `json:"install_sandbox_runs,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	InstallInputs      []InstallInputs     `json:"install_inputs" gorm:"constraint:OnDelete:CASCADE;"`

	AWSAccount AWSAccount `json:"aws_account" gorm:"constraint:OnDelete:CASCADE;"`
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
