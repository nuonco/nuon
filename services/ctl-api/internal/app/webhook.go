package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

// Webhook is an outbound webhook subscription owned by an org. The same JSONB
// `interests` filter shape used by SlackChannelSubscription gates which events
// each row receives, and the optional `match` predicate further scopes
// deliveries to specific installs / components / actions or by labels (mirror
// of SlackChannelSubscription.Match — see pkg/labels/match.go).
//
// Routing predicate:
//
//   - Match == nil → "org-wide" subscription: every event in the row's org
//     fires (subject to Interests).
//   - Match != nil → labels.SubscriptionMatch.Matches is evaluated against
//     the dispatch's labels.EventTargets in
//     internal/pkg/queue/signal/hooks/webhook.go before delivery.
//
// MatchCanonical is a generated text mirror of Match.Canonical(). It exists
// solely so the unique index on (org_id, webhook_url, match_canonical,
// deleted_at) can collapse semantically-equal Match values into the same key
// (JSONB doesn't support direct uniqueness, and Postgres treats two
// structurally-different JSONB encodings of the same logical value as
// distinct). Canonical() returns "" for nil/zero Match, so the org-wide row
// also collapses to a deterministic single-row-per-(org,url) key — preserving
// the legacy "one URL per org" invariant for unscoped webhooks while
// allowing additional rows scoped to different Match predicates.
type Webhook struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_webhooks_org_url_match" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_webhooks_org;uniqueIndex:idx_webhooks_org_url_match" temporaljson:"org_id,omitzero,omitempty"`

	WebhookURL    string `json:"webhook_url,omitzero" gorm:"notnull;uniqueIndex:idx_webhooks_org_url_match" temporaljson:"webhook_url,omitzero,omitempty"`
	WebhookSecret string `json:"-" temporaljson:"webhook_secret,omitzero,omitempty"`

	// Match is the per-webhook routing predicate. Nil = match every event in
	// the org. Non-nil = evaluated by labels.SubscriptionMatch against the
	// dispatch-time labels.EventTargets. swaggertype:"object" keeps the SDK
	// from materialising the full nested type tree (mirrors the same
	// treatment on SlackChannelSubscription.Match).
	Match *labels.SubscriptionMatch `json:"match,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"match,omitzero,omitempty"`

	// MatchCanonical mirrors Match.Canonical() so the unique index can
	// dedupe semantically-equal predicates. Maintained by BeforeSave below.
	// Internal index helper, not part of the public API surface.
	MatchCanonical string `json:"-" gorm:"type:text;not null;default:'';uniqueIndex:idx_webhooks_org_url_match" temporaljson:"-"`

	// Interests is the per-webhook event filter. Stored as JSONB; the same
	// shape is used by other notification surfaces (see internal/pkg/interests).
	// New rows default to AllEvents=true via the create handler when the
	// request omits the field. Pre-existing rows whose JSONB column is
	// NULL/empty are treated as AllEvents=true at delivery time by the
	// dispatcher in internal/pkg/queue/signal/hooks/webhook.go, so legacy
	// webhooks keep receiving every supported event without a backfill.
	Interests interests.Interests `json:"interests,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"interests,omitzero,omitempty"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

func (a *Webhook) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewWebhookID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

// BeforeSave keeps MatchCanonical in lockstep with Match. Match.Canonical()
// returns "" for nil/zero receiver — the desired default for legacy/org-wide
// rows (which all collapse to the same deterministic index key per
// (org_id, webhook_url)). Runs on both inserts and updates so an edited
// Match recomputes correctly.
func (a *Webhook) BeforeSave(tx *gorm.DB) error {
	a.MatchCanonical = a.Match.Canonical()
	return nil
}
