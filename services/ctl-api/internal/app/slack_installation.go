package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// SlackInstallationStatus reflects the lifecycle of an OAuth installation
// for a Slack workspace.
type SlackInstallationStatus string

const (
	SlackInstallationStatusActive      SlackInstallationStatus = "active"
	SlackInstallationStatusUninstalled SlackInstallationStatus = "uninstalled"
	SlackInstallationStatusDisabled    SlackInstallationStatus = "disabled"
)

// SlackInstallation represents a Slack workspace ("team") installation of the
// Nuon Slack app. It is workspace-scoped (NOT org-scoped) — a single workspace
// may be linked to one or more Nuon orgs via SlackOrgLink rows.
//
// The bot access token is stored in plaintext for now, mirroring the existing
// app.Webhook.WebhookSecret pattern. A future migration may move this behind
// an encryption helper.
type SlackInstallation struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_slack_installations_team" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// TeamID is the stable Slack workspace identifier (e.g. "T0123456789").
	// Combined with deleted_at to allow re-installation after uninstall.
	TeamID       string `json:"team_id,omitzero" gorm:"notnull;uniqueIndex:idx_slack_installations_team" temporaljson:"team_id,omitzero,omitempty"`
	TeamName     string `json:"team_name,omitzero" temporaljson:"team_name,omitzero,omitempty"`
	EnterpriseID string `json:"enterprise_id,omitzero" temporaljson:"enterprise_id,omitzero,omitempty"`

	BotUserID string `json:"bot_user_id,omitzero" temporaljson:"bot_user_id,omitzero,omitempty"`
	AppID     string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	Scope     string `json:"scope,omitzero" temporaljson:"scope,omitzero,omitempty"`

	Status SlackInstallationStatus `json:"status,omitzero" gorm:"notnull;default:'active'" temporaljson:"status,omitzero,omitempty"`

	// BotAccessToken is stored PLAINTEXT for now (mirrors app.Webhook.WebhookSecret).
	BotAccessToken string `json:"-" temporaljson:"bot_access_token,omitzero,omitempty"`

	InstalledBySlackUserID string  `json:"installed_by_slack_user_id,omitzero" temporaljson:"installed_by_slack_user_id,omitzero,omitempty"`
	InstalledByAccountID   string  `json:"installed_by_account_id,omitzero" gorm:"notnull" temporaljson:"installed_by_account_id,omitzero,omitempty"`
	InstalledByAccount     Account `json:"-" gorm:"foreignKey:InstalledByAccountID;references:ID" temporaljson:"installed_by_account,omitzero,omitempty"`
}

func (SlackInstallation) TableName() string {
	return "slack_installations"
}

func (a *SlackInstallation) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewSlackInstallationID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.Status == "" {
		a.Status = SlackInstallationStatusActive
	}

	return nil
}
