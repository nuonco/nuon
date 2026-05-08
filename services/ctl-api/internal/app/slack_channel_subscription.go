package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

// SlackChannelSubscription is a per-channel routing rule belonging to a
// SlackOrgLink. The CASCADE FK from OrgLinkID structurally enforces the trust
// binding: if the workspace ↔ org link is hard-deleted, all of its channel
// subscriptions disappear with it.
//
// Routing predicate:
//
//   - Match == nil  → "org-wide" subscription, fires for every event in the
//     linked org (subject to the per-row Interests filter).
//   - Match != nil  → labels.SubscriptionMatch.Matches is evaluated against
//     the dispatch's labels.EventTargets. See pkg/labels/match.go for the
//     composition rules (OR across populated kinds, OR within a TargetMatch
//     between IDs and Selector, AND across Selector.MatchLabels keys).
//
// MatchCanonical is a generated text mirror of Match.Canonical(). It exists
// solely so the unique index on (team_id, channel_id, org_link_id,
// match_canonical, deleted_at) can collapse semantically-equal Match values
// into the same key — JSONB doesn't support direct uniqueness, and Postgres
// treats two structurally-different JSONB encodings of the same logical
// value as distinct. Canonical() returns "" for nil/zero Match, so the
// org-wide row also collapses to a deterministic single-row-per-channel key.
//
// At least one of CreatedBySlackUserID / CreatedByAccountID must be populated;
// this is enforced via migration 101 (GORM doesn't model multi-column CHECKs
// cleanly).
type SlackChannelSubscription struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_slack_channel_subs_team_channel_link_match" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgLinkID string       `json:"org_link_id,omitzero" gorm:"notnull;index:idx_slack_channel_subs_link;uniqueIndex:idx_slack_channel_subs_team_channel_link_match" temporaljson:"org_link_id,omitzero,omitempty"`
	OrgLink   SlackOrgLink `json:"-" gorm:"foreignKey:OrgLinkID;references:ID;constraint:OnDelete:CASCADE" temporaljson:"org_link,omitzero,omitempty"`

	TeamID      string `json:"team_id,omitzero" gorm:"notnull;index:idx_slack_channel_subs_team;uniqueIndex:idx_slack_channel_subs_team_channel_link_match" temporaljson:"team_id,omitzero,omitempty"`
	ChannelID   string `json:"channel_id,omitzero" gorm:"notnull;uniqueIndex:idx_slack_channel_subs_team_channel_link_match" temporaljson:"channel_id,omitzero,omitempty"`
	ChannelName string `json:"channel_name,omitzero" temporaljson:"channel_name,omitzero,omitempty"`

	// OrgID is denormalized from OrgLink for query convenience.
	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_slack_channel_subs_org" temporaljson:"org_id,omitzero,omitempty"`

	// Match is the per-subscription routing predicate. Nil = match every
	// event in the org. Non-nil = evaluated by labels.SubscriptionMatch
	// against the dispatch-time labels.EventTargets. swaggertype:"object"
	// keeps the SDK from materialising the full nested type tree.
	Match *labels.SubscriptionMatch `json:"match,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"match,omitzero,omitempty"`

	// MatchCanonical mirrors Match.Canonical() so the unique index can
	// dedupe semantically-equal predicates. Maintained by BeforeSave below.
	// Internal index helper, not part of the public API surface.
	MatchCanonical string `json:"-" gorm:"type:text;not null;default:'';uniqueIndex:idx_slack_channel_subs_team_channel_link_match" temporaljson:"-"`

	// Interests is the per-subscription event filter. Stored as JSONB; one
	// shape is shared with webhooks (see internal/pkg/interests). New rows
	// default to AllEvents=true via the create handler when the request omits
	// the field.
	Interests interests.Interests `json:"interests,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"interests,omitzero,omitempty"`

	CreatedBySlackUserID *string  `json:"created_by_slack_user_id,omitzero" temporaljson:"created_by_slack_user_id,omitzero,omitempty"`
	CreatedByAccountID   *string  `json:"created_by_account_id,omitzero" temporaljson:"created_by_account_id,omitzero,omitempty"`
	CreatedByAccount     *Account `json:"-" gorm:"foreignKey:CreatedByAccountID;references:ID" temporaljson:"created_by_account,omitzero,omitempty"`
}

func (SlackChannelSubscription) TableName() string {
	return "slack_channel_subscriptions"
}

func (a *SlackChannelSubscription) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewSlackChannelSubscriptionID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

// BeforeSave keeps MatchCanonical in lockstep with Match. Match.Canonical()
// returns "" for nil/zero receiver — the desired default for org-wide rows
// (which all collapse to the same deterministic index key per
// (team, channel, link)). Runs on both inserts and updates so an edited
// Match recomputes correctly.
func (a *SlackChannelSubscription) BeforeSave(tx *gorm.DB) error {
	a.MatchCanonical = a.Match.Canonical()
	return nil
}
