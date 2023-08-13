package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type SandboxRelease struct {
	Model

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
