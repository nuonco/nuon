package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
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
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`

	Status            AppConfigStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description,omitzero" gorm:"notnull;default null" temporaljson:"status_description,omitzero,omitempty"`

	State        string         `json:"state,omitzero" temporaljson:"state,omitzero,omitempty"`
	Readme       string         `json:"readme,omitzero" temporaljson:"readme,omitzero,omitempty"`
	Checksum     string         `json:"checksum,omitzero" temporaljson:"checksum,omitzero,omitempty"`
	CLIVersion   string         `json:"cli_version,omitzero" gorm:"default null" temporaljson:"cli_version,omitzero,omitempty"`
	ComponentIDs pq.StringArray `gorm:"type:text[]" json:"component_ids,omitzero" temporaljson:"component_ids,omitzero,omitempty" swaggertype:"array,string"`

	// Lookups on the app config

	PermissionsConfig          AppPermissionsConfig        `json:"permissions,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"permissions_config,omitzero,omitempty"`
	BreakGlassConfig           AppBreakGlassConfig         `json:"break_glass,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"break_glass_config,omitzero,omitempty"`
	PoliciesConfig             AppPoliciesConfig           `json:"policies,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"policies_config,omitzero,omitempty"`
	SecretsConfig              AppSecretsConfig            `json:"secrets,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"secrets_config,omitzero,omitempty"`
	SandboxConfig              AppSandboxConfig            `json:"sandbox,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"sandbox_config,omitzero,omitempty"`
	InputConfig                AppInputConfig              `json:"input,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"input_config,omitzero,omitempty"`
	RunnerConfig               AppRunnerConfig             `json:"runner,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"runner_config,omitzero,omitempty"`
	StackConfig                AppStackConfig              `json:"stack,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"stack_config,omitzero,omitempty"`
	ComponentConfigConnections []ComponentConfigConnection `json:"component_config_connections,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_config_connections,omitzero,omitempty"`
	ActionWorkflowConfigs      []ActionWorkflowConfig      `json:"action_workflow_configs,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"action_workflow_configs,omitzero,omitempty"`

	// individual pointers
	InstallAWSCloudFormationStackVersion []InstallStackVersion `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_aws_cloud_formation_stack_version,omitzero,omitempty"`

	// fields that are filled in via after query or views
	Version int `json:"version,omitzero" gorm:"->;-:migration" temporaljson:"version,omitzero,omitempty"`

	AppBranchID generics.NullString `json:"app_branch_id,omitzero" gorm:"index:idx_app_app_branch" swaggertype:"string" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   *AppBranch          `json:"app_branch" temporaljson:"app_branch,omitzero,omitempty"`

	VCSConnectionCommitID generics.NullString  `json:"-"  swaggertype:"string" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`
}

func (a *AppConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
			AlwaysReapply: false,
		},
		{
			Name:          views.CustomViewName(db, &AppConfig{}, "latest_view_v1"),
			SQL:           viewsql.AppConfigsLatestViewV1,
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
