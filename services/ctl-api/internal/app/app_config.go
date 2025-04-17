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

type AppConfigType string

const (
	AppConfigTypeToml   AppConfigType = "toml"
	AppConfigTypeManual AppConfigType = "manual"
)

type AppConfigVersion string

const (
	AppConfigVersionDefault AppConfigVersion = ""
	AppConfigVersionV2      AppConfigVersion = "v2"
)

type AppConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppID string `json:"app_id"`

	Status            AppConfigStatus `json:"status"`
	StatusDescription string          `json:"status_description" gorm:"notnull;default null"`

	State    string `json:"state"`
	Readme   string `json:"readme"`
	Checksum string `json:"checksum"`

	// Lookups on the app config
	PermissionsConfig          AppPermissionsConfig        `json:"permissions,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	BreakGlassConfig           AppBreakGlassConfig         `json:"break_glass,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	PoliciesConfig             AppPoliciesConfig           `json:"policies,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	SecretsConfig              AppSecretsConfig            `json:"secrets,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	SandboxConfig              AppSandboxConfig            `json:"sandbox,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	InputConfig                AppInputConfig              `json:"input,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	RunnerConfig               AppRunnerConfig             `json:"runner,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	CloudFormationStackConfig  AppStackConfig              `json:"cloudformation_stack,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	ComponentConfigConnections []ComponentConfigConnection `json:"component_config_connections,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	// individual pointers
	InstallAWSCloudFormationStackVersion []InstallStackVersion `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	// fields that are filled in via after query or views
	Version int `json:"version" gorm:"->;-:migration"`
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
