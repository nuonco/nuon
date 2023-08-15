package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type SandboxRelease struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	SandboxID               string  `json:"-"`
	Sandbox                 Sandbox `json:"-"`
	Version                 string  `gorm:"unique" json:"version"`
	TerraformVersion        string  `json:"terraform_version"`
	ProvisionPolicyURL      string  `json:"provision_policy_url"`
	DeprovisionPolicyURL    string  `json:"deprovision_policy_url"`
	TrustPolicyURL          string  `json:"trust_policy_url"`
	OneClickRoleTemplateURL string  `json:"one_click_role_template_url"`
}

func (s *SandboxRelease) BeforeCreate(tx *gorm.DB) error {
	s.ID = domains.NewSandboxReleaseID()
	return nil
}
