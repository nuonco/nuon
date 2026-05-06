package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

type Webhook struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_webhooks_org_url" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_webhooks_org;uniqueIndex:idx_webhooks_org_url" temporaljson:"org_id,omitzero,omitempty"`

	WebhookURL    string `json:"webhook_url,omitzero" gorm:"notnull;uniqueIndex:idx_webhooks_org_url" temporaljson:"webhook_url,omitzero,omitempty"`
	WebhookSecret string `json:"-" temporaljson:"webhook_secret,omitzero,omitempty"`

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
