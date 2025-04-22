package app

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TerraformState struct {
	ID        string    `gorm:"column:state_id;type:text;not null;primaryKey;index:idx_state_id"`
	Revision  int       `gorm:"column:revision;not null;primaryKey"`
	Data      []byte    `gorm:"column:data;type:bytea"`
	Lock      []byte    `gorm:"column:lock;type:bytea"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (s *TerraformState) BeforeCreate(tx *gorm.DB) (err error) {
	if s.Revision == 0 {
		s.Revision = 1
	}

	return nil
}

type TerraformStateData struct {
	Version          int            `json:"version"`
	TerraformVersion string         `json:"terraform_version"`
	Serial           int            `json:"serial"`
	Lineage          string         `json:"lineage"`
	Outputs          map[string]any `json:"outputs"`
	Resources        []Resource     `json:"resources"`
	CheckResults     any            `json:"check_results"`
}

type TerraformWorkspace struct {
	ID        string                  `gorm:"column:id;type:text;not null;index:idx_workspace_id"`
	OrgID     string                  `gorm:"column:org_id;type:text;not null;index:idx_org_owner,unique"`
	OwnerID   string                  `gorm:"column:owner_id;type:text;not null;index:idx_org_owner,unique"`
	OwnerType TerraformWorkspaceOwner `gorm:"column:owner_type;type:text;not null;index:idx_org_owner,unique"`
	CreatedAt time.Time               `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time               `gorm:"column:updated_at;autoUpdateTime"`
}

type TerraformWorkspaceOwner string

const (
	TerraformWorkspaceOwnerInstallSandboxRun TerraformWorkspaceOwner = "install_sandbox_run"
	TerraformWorkspaceOwnerInstallComponent  TerraformWorkspaceOwner = "install_component"
)

func (s *TerraformWorkspace) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type Resource struct {
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Provider  string     `json:"provider"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	SchemaVersion       int            `json:"schema_version"`
	Attributes          map[string]any `json:"attributes"`
	SensitiveAttributes []any          `json:"sensitive_attributes"`
}

// Lock a lock on state
type TerraformLock struct {
	Created   string
	Path      string
	ID        string
	Operation string
	Info      string
	Who       string
	Version   any
}
