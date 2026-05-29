package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

// DatadogEventSubscription is a per-connection routing rule. It mirrors
// app.SlackChannelSubscription one-to-one in shape (Match + Interests +
// MatchCanonical for dedup), minus the channel/team fields — DD events go
// to the connection's tenant event stream rather than a named channel.
//
// Routing predicate:
//
//   - Match == nil  → "org-wide" subscription, fires for every event in
//     the connection's org (subject to the per-row Interests filter).
//   - Match != nil  → labels.SubscriptionMatch.Matches is evaluated
//     against the dispatch's labels.EventTargets. See pkg/labels/match.go.
//
// MatchCanonical exists for the same reason as on SlackChannelSubscription:
// JSONB doesn't support direct uniqueness, so we mirror Match.Canonical()
// into a text column and put the unique index on it.
//
// AdditionalTags are appended onto the connection's DefaultTags when an
// event fires through this subscription — typical use: scope the
// subscription to a customer install AND tag every emit with
// customer:acme so monitors can filter on it.
//
// AlertTypeOverride / PriorityOverride let a subscription force a specific
// DD alert level instead of the hook's default mapping. Empty → use
// default. Validated at create/update time, not in BeforeSave (so the DB
// stays permissive of legacy rows).
type DatadogEventSubscription struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_datadog_event_subs_conn_match" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	ConnectionID string            `json:"connection_id,omitzero" gorm:"notnull;index:idx_datadog_event_subs_conn;uniqueIndex:idx_datadog_event_subs_conn_match" temporaljson:"connection_id,omitzero,omitempty"`
	Connection   DatadogConnection `json:"-" gorm:"foreignKey:ConnectionID;references:ID;constraint:OnDelete:CASCADE" temporaljson:"connection,omitzero,omitempty"`

	// OrgID is denormalized from Connection for query convenience —
	// matches the SlackChannelSubscription pattern so the lifecycle hook
	// can index by (org_id, connection_id) without a join.
	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_datadog_event_subs_org" temporaljson:"org_id,omitzero,omitempty"`

	Match          *labels.SubscriptionMatch `json:"match,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"match,omitzero,omitempty"`
	MatchCanonical string                    `json:"-" gorm:"type:text;not null;default:'';uniqueIndex:idx_datadog_event_subs_conn_match" temporaljson:"-"`

	Interests interests.Interests `json:"interests,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"interests,omitzero,omitempty"`

	AdditionalTags pq.StringArray `json:"additional_tags,omitzero" gorm:"type:text[]" swaggertype:"array,string" temporaljson:"additional_tags,omitzero,omitempty"`

	// AlertTypeOverride forces a DD alert_type — one of "info", "warning",
	// "error", "success". Empty leaves the hook's default mapping in
	// place. Stored as a plain string so DD's accepted values can evolve
	// without a schema change.
	AlertTypeOverride string `json:"alert_type_override,omitzero" temporaljson:"alert_type_override,omitzero,omitempty"`

	// PriorityOverride forces a DD priority — "normal" or "low". Empty
	// leaves the hook's default in place.
	PriorityOverride string `json:"priority_override,omitzero" temporaljson:"priority_override,omitzero,omitempty"`
}

func (DatadogEventSubscription) TableName() string {
	return "datadog_event_subscriptions"
}

func (a *DatadogEventSubscription) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewDatadogEventSubscriptionID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

// BeforeSave keeps MatchCanonical in lockstep with Match. Match.Canonical()
// returns "" for nil/zero receiver — the desired default for org-wide rows
// (which all collapse to the same deterministic index key per connection).
func (a *DatadogEventSubscription) BeforeSave(tx *gorm.DB) error {
	a.MatchCanonical = a.Match.Canonical()
	return nil
}
