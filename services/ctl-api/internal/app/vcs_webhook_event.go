package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"gorm.io/gorm"
)

type VCSWebhookEventStatus string

const (
	VCSWebhookEventStatusPending VCSWebhookEventStatus = "pending"
	VCSWebhookEventStatusEmitted VCSWebhookEventStatus = "emitted"
)

type VCSWebhookEvent struct {
	ID        string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedAt time.Time `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt time.Time `json:"updated_at,omitzero" gorm:"notnull"`

	EventType string          `json:"event_type" gorm:"not null;default:''"`
	Body      *blobstore.Blob `json:"body" gorm:"type:jsonb;default:null"`
	Status    string          `json:"status" gorm:"not null;default:'pending'"`
}

func (v *VCSWebhookEvent) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSWebhookEventID()
	if v.Status == "" {
		v.Status = string(VCSWebhookEventStatusPending)
	}
	return nil
}

// ParseBody loads the blob content and parses it into a map for processing.
func (v *VCSWebhookEvent) ParseBody(ctx context.Context) (map[string]any, error) {
	if v.Body == nil {
		return nil, fmt.Errorf("webhook event body is nil")
	}

	raw, err := v.Body.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load webhook body blob: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal webhook body: %w", err)
	}
	return payload, nil
}
