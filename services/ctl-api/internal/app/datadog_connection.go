package app

import (
	"net/url"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// DatadogConnectionStatus reflects whether a Datadog tenant binding is
// currently usable. Revoked entries are kept so the audit trail of which
// connections fanned out which events is preserved.
type DatadogConnectionStatus string

const (
	DatadogConnectionStatusVerified DatadogConnectionStatus = "verified"
	DatadogConnectionStatusRevoked  DatadogConnectionStatus = "revoked"
)

// DatadogConnectionPurpose is a UI hint distinguishing a vendor's own DD
// tenant from a customer's DD tenant. Purely cosmetic — no backend behavior
// is gated on it.
type DatadogConnectionPurpose string

const (
	DatadogConnectionPurposeInternal DatadogConnectionPurpose = "internal"
	DatadogConnectionPurposeCustomer DatadogConnectionPurpose = "customer"
)

// Known Datadog regional sites. Storage accepts either one of these keys
// OR a full https URL (for on-prem / private DD). The DD client resolves
// the API host from the stored value via resolveSiteURL.
const (
	DatadogSiteUS1 = "us1"
	DatadogSiteUS3 = "us3"
	DatadogSiteUS5 = "us5"
	DatadogSiteEU1 = "eu1"
	DatadogSiteAP1 = "ap1"
	DatadogSiteGov = "gov"
)

// DatadogConnection is a per-org binding to a Datadog tenant. An org may
// have N connections (vendor's own DD + one per customer's DD). The signal
// lifecycle hook iterates all verified connections for the event's org and
// fans events out to each connection whose subscriptions match.
//
// Site holds either a known key (us1/us3/us5/eu1/ap1/gov) or a full
// https://... URL for self-hosted DD. Validation is enforced in
// validateDatadogSite below; the DD client resolves both shapes to a base
// URL via resolveSiteURL.
//
// API and application keys are stored PLAINTEXT today, matching the
// existing app.Webhook.WebhookSecret / app.SlackInstallation.BotAccessToken
// pattern. A future migration may move them behind an encryption helper.
// ApplicationKey is optional for emit-only flows but REQUIRED for the
// one-click managed-monitor feature (DD's Monitors API needs both keys).
type DatadogConnection struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_datadog_connections_deleted" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_datadog_connections_org" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE" temporaljson:"org,omitzero,omitempty"`

	// Name is the user-facing label shown in the dashboard list — e.g.
	// "Internal monitoring" or "ACME Corp prod". Required; not unique
	// (two connections can share a label, the ID is the stable key).
	Name string `json:"name,omitzero" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`

	// Purpose is a cosmetic badge in the dashboard. Defaults to
	// "internal" via BeforeCreate when omitted.
	Purpose DatadogConnectionPurpose `json:"purpose,omitzero" gorm:"notnull;default:'internal'" temporaljson:"purpose,omitzero,omitempty"`

	// Site is either a known key (us1/us3/us5/eu1/ap1/gov) or a full
	// https://... URL. Validated in validateDatadogSite.
	Site string `json:"site,omitzero" gorm:"notnull" temporaljson:"site,omitzero,omitempty"`

	// APIKey is the DD API key used for the Events API. Required.
	// Stored plaintext.
	APIKey string `json:"-" gorm:"notnull" temporaljson:"api_key,omitzero,omitempty"`

	// ApplicationKey is the DD app key used for the Monitors API.
	// Optional — when absent, one-click managed-monitor creation is
	// disabled and the dashboard surfaces a prompt to add one.
	ApplicationKey string `json:"-" temporaljson:"application_key,omitzero,omitempty"`

	Status DatadogConnectionStatus `json:"status,omitzero" gorm:"notnull;default:'verified';index:idx_datadog_connections_org_status,composite:org_status" temporaljson:"status,omitzero,omitempty"`

	// DefaultTags are appended to every event emitted via this
	// connection, before per-subscription AdditionalTags. Stored as a
	// Postgres text[] so DD-shaped "key:value" strings survive round-trips
	// without JSON quoting noise.
	DefaultTags pq.StringArray `json:"default_tags,omitzero" gorm:"type:text[]" swaggertype:"array,string" temporaljson:"default_tags,omitzero,omitempty"`

	// DefaultNotifyHandles are DD-style @-mentions (e.g. "@pagerduty-prod",
	// "@slack-oncall") that the one-click monitor creator splices into
	// the monitor body. Per-click overrides are allowed at monitor
	// creation time. Stored as text[] for the same reason as DefaultTags.
	DefaultNotifyHandles pq.StringArray `json:"default_notify_handles,omitzero" gorm:"type:text[]" swaggertype:"array,string" temporaljson:"default_notify_handles,omitzero,omitempty"`
}

func (DatadogConnection) TableName() string {
	return "datadog_connections"
}

func (a *DatadogConnection) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewDatadogConnectionID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.Status == "" {
		a.Status = DatadogConnectionStatusVerified
	}

	if a.Purpose == "" {
		a.Purpose = DatadogConnectionPurposeInternal
	}

	return nil
}

func (a *DatadogConnection) BeforeSave(tx *gorm.DB) error {
	a.Site = strings.TrimSpace(a.Site)
	a.Name = strings.TrimSpace(a.Name)
	return validateDatadogSite(a.Site)
}

// validateDatadogSite enforces the two-shape contract: either a known
// regional key OR a full https URL with a host and no path. Anything
// else is rejected so a typo can't silently route events into the void.
func validateDatadogSite(site string) error {
	if site == "" {
		return errInvalidDatadogSite("site is required")
	}
	switch site {
	case DatadogSiteUS1, DatadogSiteUS3, DatadogSiteUS5,
		DatadogSiteEU1, DatadogSiteAP1, DatadogSiteGov:
		return nil
	}
	u, err := url.Parse(site)
	if err != nil {
		return errInvalidDatadogSite("site is not a valid URL: " + err.Error())
	}
	if u.Scheme != "https" {
		return errInvalidDatadogSite("custom site URL must be https")
	}
	if u.Host == "" {
		return errInvalidDatadogSite("custom site URL must include a host")
	}
	if u.Path != "" && u.Path != "/" {
		return errInvalidDatadogSite("custom site URL must not include a path")
	}
	return nil
}

type datadogSiteError struct{ msg string }

func (e *datadogSiteError) Error() string { return "invalid datadog site: " + e.msg }

func errInvalidDatadogSite(msg string) error { return &datadogSiteError{msg: msg} }
