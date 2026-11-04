package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallApprovalOption string

const (
	InstallApprovalOptionApproveAll InstallApprovalOption = "approve-all"
	InstallApprovalOptionPrompt     InstallApprovalOption = "prompt"
)

type InstallConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID       string                `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org                   `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `json:"-" temporaljson:"install,omitzero,omitempty"`

	ApprovalOption   InstallApprovalOption `json:"approval_option,omitzero" gorm:"not null;default:'prompt'" temporaljson:"approval_option,omitzero,omitempty"`
	TelemetryEnabled *bool                 `json:"telemetry_enabled" temporaljson:"telemetry_enabled,omitempty" extensions:"x-nullable"`

	// Per-install stack template overrides (nil = use app config default)
	VPCNestedTemplateURL     *string                               `json:"vpc_nested_template_url,omitempty" gorm:"column:vpc_nested_template_url" temporaljson:"vpc_nested_template_url,omitempty"`
	RunnerNestedTemplateURL  *string                               `json:"runner_nested_template_url,omitempty" gorm:"column:runner_nested_template_url" temporaljson:"runner_nested_template_url,omitempty"`
	UIDefinition             *config.UploadedUIDefinition          `json:"ui_definition,omitempty" gorm:"column:uidefinition;type:jsonb;serializer:json" temporaljson:"ui_definition,omitzero,omitempty"`
	CustomNestedStacks       []config.CustomNestedStack            `json:"custom_nested_stacks,omitempty" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"custom_nested_stacks,omitempty"`
	AWSStackParameters       map[string]string                     `json:"-" gorm:"type:jsonb;serializer:json" temporaljson:"aws_stack_parameters,omitempty"`
	AWSNestedStackParameters map[string]config.CloudFormationStack `json:"-" gorm:"type:jsonb;serializer:json" temporaljson:"aws_nested_stack_parameters,omitempty"`
	ComponentToggles         map[string]bool                       `json:"component_toggles,omitempty" gorm:"type:jsonb;serializer:json;default:'{}'" temporaljson:"component_toggles,omitempty"`
	RunnerImageAutoUpdate    *bool                                 `json:"runner_image_auto_update,omitempty" gorm:"default null" temporaljson:"runner_image_auto_update,omitempty" extensions:"x-nullable"`
	RunnerImageID            *string                               `json:"runner_image_id,omitempty" gorm:"default null" temporaljson:"runner_image_id,omitempty"`
	ActionTimeout            *time.Duration                        `json:"action_timeout,omitempty" gorm:"default null" swaggertype:"primitive,integer" temporaljson:"action_timeout,omitempty" extensions:"x-nullable"`
	labels.Labeled
}

func (c *InstallConfig) ResolveActionTimeout(configured time.Duration) time.Duration {
	if c != nil && c.ActionTimeout != nil {
		return min(max(*c.ActionTimeout, config.MinActionTimeout), config.MaxActionTimeout)
	}
	return min(configured, DefaultMaxActionTimeout)
}

func (a *InstallConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallConfigID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallConfig) IsComponentEnabled(componentID string, ccc *ComponentConfigConnection) bool {
	if ccc == nil || !ccc.IsToggleable() {
		return true
	}
	if c != nil && c.ComponentToggles != nil {
		if enabled, ok := c.ComponentToggles[componentID]; ok {
			return enabled
		}
	}
	return ccc.GetDefaultEnabled()
}

func (c *InstallConfig) IsTelemetryEnabled(orgDefault bool) bool {
	if c != nil && c.TelemetryEnabled != nil {
		return *c.TelemetryEnabled
	}
	return orgDefault
}

func (c *InstallConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallConfig{}, "uq"),
			Columns: []string{
				"install_id",
				"deleted_at",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name: indexes.Name(db, &InstallConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
