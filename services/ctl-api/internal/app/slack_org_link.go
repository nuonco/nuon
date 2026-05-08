package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// SlackOrgLinkStatus reflects whether a workspace ↔ org binding is currently
// trusted.
type SlackOrgLinkStatus string

const (
	SlackOrgLinkStatusVerified SlackOrgLinkStatus = "verified"
	SlackOrgLinkStatusRevoked  SlackOrgLinkStatus = "revoked"
)

// SlackOrgLink is the trust binding between a Slack workspace (TeamID) and a
// Nuon org (OrgID). A message lands in workspace T for org O iff:
//
//	(a) installation T is active,
//	(b) org_link (T, O) is verified,
//	(c) channel sub (T, channel, O) is active.
//
// The FK from sub → link with ON DELETE CASCADE makes (b) structural for
// channel subscriptions.
//
// NOTE: a strict PG FK from TeamID → slack_installations.team_id is not
// declared here because the parent table uses soft-delete (uniqueness on
// (team_id, deleted_at)), which is incompatible with PG's FK requirement of a
// non-partial UNIQUE constraint on the referenced column. The
// installation lifecycle is enforced at the application layer via the
// installation Status field and uninstall event handling.
type SlackOrgLink struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_slack_org_links_team_org" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	TeamID string `json:"team_id,omitzero" gorm:"notnull;index:idx_slack_org_links_team;uniqueIndex:idx_slack_org_links_team_org" temporaljson:"team_id,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_slack_org_links_org;uniqueIndex:idx_slack_org_links_team_org" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE" temporaljson:"org,omitzero,omitempty"`

	Status SlackOrgLinkStatus `json:"status,omitzero" gorm:"notnull;default:'verified'" temporaljson:"status,omitzero,omitempty"`

	LinkedByAccountID string  `json:"linked_by_account_id,omitzero" gorm:"notnull" temporaljson:"linked_by_account_id,omitzero,omitempty"`
	LinkedByAccount   Account `json:"-" gorm:"foreignKey:LinkedByAccountID;references:ID" temporaljson:"linked_by_account,omitzero,omitempty"`
}

func (SlackOrgLink) TableName() string {
	return "slack_org_links"
}

func (a *SlackOrgLink) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewSlackOrgLinkID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.Status == "" {
		a.Status = SlackOrgLinkStatusVerified
	}

	return nil
}
