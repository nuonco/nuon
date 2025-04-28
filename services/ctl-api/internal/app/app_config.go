package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type AppConfigStatus string

const (
	AppConfigStatusActive   AppConfigStatus = "active"
	AppConfigStatusPending  AppConfigStatus = "pending"
	AppConfigStatusSyncing  AppConfigStatus = "syncing"
	AppConfigStatusError    AppConfigStatus = "error"
	AppConfigStatusOutdated AppConfigStatus = "outdated"
)

// type AppConfigType string

// const (
// 	AppConfigTypeToml   AppConfigType = "toml"
// 	AppConfigTypeManual AppConfigType = "manual"
// )

type AppConfigVersion string

const (
	AppConfigVersionDefault AppConfigVersion = ""
	AppConfigVersionV2      AppConfigVersion = "v2"
)

type AppConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id" temporaljson:"app_id,omitzero,omitempty"`

	Status            AppConfigStatus `json:"status" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description" gorm:"notnull;default null" temporaljson:"status_description,omitzero,omitempty"`

	State    string `json:"state" temporaljson:"state,omitzero,omitempty"`
	Readme   string `json:"readme" temporaljson:"readme,omitzero,omitempty"`
	Checksum string `json:"checksum" temporaljson:"checksum,omitzero,omitempty"`

	// Lookups on the app config

	PermissionsConfig          AppPermissionsConfig        `json:"permissions,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"permissions_config,omitzero,omitempty"`
	BreakGlassConfig           AppBreakGlassConfig         `json:"break_glass,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"break_glass_config,omitzero,omitempty"`
	PoliciesConfig             AppPoliciesConfig           `json:"policies,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"policies_config,omitzero,omitempty"`
	SecretsConfig              AppSecretsConfig            `json:"secrets,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"secrets_config,omitzero,omitempty"`
	SandboxConfig              AppSandboxConfig            `json:"sandbox,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"sandbox_config,omitzero,omitempty"`
	InputConfig                AppInputConfig              `json:"input,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"input_config,omitzero,omitempty"`
	RunnerConfig               AppRunnerConfig             `json:"runner,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"runner_config,omitzero,omitempty"`
	StackConfig                AppStackConfig              `json:"stack,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"stack_config,omitzero,omitempty"`
	ComponentConfigConnections []ComponentConfigConnection `json:"component_config_connections,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_config_connections,omitzero,omitempty"`

	// individual pointers
	InstallAWSCloudFormationStackVersion []InstallStackVersion `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_aws_cloud_formation_stack_version,omitzero,omitempty"`

	// fields that are filled in via after query or views
	Version int `json:"version" gorm:"->;-:migration" temporaljson:"version,omitzero,omitempty"`
}

func (a AppConfig) UseView() bool {
	return true
}

func (a AppConfig) ViewVersion() string {
	return "v2"
}

func (i *AppConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &AppConfig{}, 2),
			SQL:           viewsql.AppConfigViewV2,
			AlwaysReapply: true,
		},
	}
}

func (a *AppConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
