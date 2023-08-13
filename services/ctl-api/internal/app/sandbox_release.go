package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type SandboxRelease struct {
	Model

	SandboxID               string
	Version                 string `gorm:"unique"`
	TerraformVersion        string
	ProvisionPolicyURL      string
	DeprovisionPolicyURL    string
	TrustPolicyURL          string
	OneClickRoleTemplateURL string
}

func (s *SandboxRelease) BeforeCreate(tx *gorm.DB) error {
	s.ID = domains.NewSandboxReleaseID()
	return nil
}
