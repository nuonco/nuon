package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

var AllPolicyTypes []config.AppPolicyType = []config.AppPolicyType{
	config.AppPolicyTypeKubernetesClusterKyverno,
	config.AppPolicyTypeTerraformDeployRunnerJobKyverno,
	config.AppPolicyTypeHelmDeployRunnerJobKyverno,
	config.AppPolicyTypeActionWorkflowRunnerJobKyverno,
}

type AppPolicyConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID       string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	AppID       string `json:"app_id,omitzero" gorm:"notnull;default null" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" gorm:"notnull;default null" temporaljson:"app_config_id,omitzero,omitempty"`

	AppPoliciesConfigID string            `json:"app_policies_config,omitzero" gorm:"notnull;default null" temporaljson:"app_policies_config_id,omitzero,omitempty"`
	AppPoliciesConfig   AppPoliciesConfig `json:"-" temporaljson:"app_policies_config,omitzero,omitempty"`

	Type     config.AppPolicyType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	Contents string               `json:"contents,omitzero" features:"template" temporaljson:"contents,omitzero,omitempty"`
}

func (a *AppPolicyConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppPolicyConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AppPolicyConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
