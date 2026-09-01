package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppBranchRunPreview struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchRunID string       `json:"app_branch_run_id,omitzero" gorm:"not null;uniqueIndex" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"-" temporaljson:"app_branch_run,omitzero,omitempty"`

	Source AppBranchRunPreviewSource `json:"source,omitzero" gorm:"not null" temporaljson:"source,omitzero,omitempty"`
	Mode   AppBranchRunPreviewMode   `json:"mode,omitzero" gorm:"not null" temporaljson:"mode,omitzero,omitempty"`

	InstallID   string `json:"install_id,omitempty" temporaljson:"install_id,omitzero,omitempty"`
	InstallName string `json:"install_name,omitempty" temporaljson:"install_name,omitzero,omitempty"`

	GitRef           string `json:"git_ref,omitempty" temporaljson:"git_ref,omitzero,omitempty"`
	InputAppConfigID string `json:"input_app_config_id,omitempty" temporaljson:"input_app_config_id,omitzero,omitempty"`

	BranchPreviewConfig   AppBranchPreviewConfig    `json:"branch_preview_config,omitzero" gorm:"type:jsonb;serializer:json;not null" temporaljson:"branch_preview_config,omitzero,omitempty"`
	OverridePreviewConfig *AppBranchPreviewOverride `json:"override_preview_config,omitempty" gorm:"type:jsonb;serializer:json;default:null" temporaljson:"override_preview_config,omitzero,omitempty"`
	ResolvedPreviewConfig AppBranchPreviewConfig    `json:"resolved_preview_config,omitzero" gorm:"type:jsonb;serializer:json;not null" temporaljson:"resolved_preview_config,omitzero,omitempty"`

	IgnoreChangesRegex   string `json:"ignore_changes_regex,omitempty" temporaljson:"ignore_changes_regex,omitzero,omitempty"`
	SendStatusesOnIgnore bool   `json:"send_statuses_on_ignore,omitempty" temporaljson:"send_statuses_on_ignore,omitzero,omitempty"`
}

func (p *AppBranchRunPreview) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranchRunPreview{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchRunPreview{}, "app_branch_run_id"),
			Columns: []string{
				"app_branch_run_id",
			},
		},
	}
}

func (p *AppBranchRunPreview) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = domains.NewAppBranchRunPreviewID()
	}
	if p.CreatedByID == "" {
		p.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if p.OrgID == "" {
		p.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (p *AppBranchRunPreview) GitHubSetStatuses() bool {
	if p == nil {
		return false
	}
	return p.ResolvedPreviewConfig.SetStatuses
}

func (p *AppBranchRunPreview) GitHubComment() bool {
	if p == nil {
		return false
	}
	return p.ResolvedPreviewConfig.Comment
}
