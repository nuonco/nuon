package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type SandboxRelease struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	SandboxID               string  `json:"-" gorm:"index:idx_sandbox_release,unique"`
	Sandbox                 Sandbox `json:"-"`
	Version                 string  `gorm:"index:idx_sandbox_release,unique" json:"version"`
	ProvisionPolicyURL      string  `json:"provision_policy_url" gorm:"notnull"`
	DeprovisionPolicyURL    string  `json:"deprovision_policy_url" gorm:"notnull"`
	TrustPolicyURL          string  `json:"trust_policy_url" gorm:"notnull"`
	OneClickRoleTemplateURL string  `json:"one_click_role_template_url" gorm:"notnull"`
}

func (s *SandboxRelease) BeforeCreate(tx *gorm.DB) error {
	s.ID = domains.NewSandboxReleaseID()
	s.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
