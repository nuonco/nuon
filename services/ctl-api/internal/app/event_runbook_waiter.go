package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type EventRunbookWaiterStatus string

const (
	EventRunbookWaiterStatusActive    EventRunbookWaiterStatus = "active"
	EventRunbookWaiterStatusMatched   EventRunbookWaiterStatus = "matched"
	EventRunbookWaiterStatusCancelled EventRunbookWaiterStatus = "cancelled"
	EventRunbookWaiterStatusExpired   EventRunbookWaiterStatus = "expired"
)

type EventRunbookWaiter struct {
	ID             string                   `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
	OrgID          string                   `gorm:"not null;<-:create" json:"org_id"`
	AppID          string                   `gorm:"not null;<-:create" json:"app_id"`
	InstallID      string                   `gorm:"not null;<-:create" json:"install_id"`
	WorkflowID     string                   `gorm:"not null;<-:create" json:"workflow_id"`
	WorkflowStepID string                   `gorm:"not null;uniqueIndex;<-:create" json:"workflow_step_id"`
	QueueSignalID  string                   `gorm:"not null;<-:create" json:"queue_signal_id"`
	TriggerID      string                   `gorm:"not null;<-:create" json:"trigger_id"`
	EventTypes     pq.StringArray           `gorm:"type:text[];<-:create" json:"event_types"`
	Filters        []TriggerFilter          `gorm:"serializer:json;type:jsonb;<-:create" json:"filters"`
	Status         EventRunbookWaiterStatus `gorm:"not null;check:event_runbook_waiter_status_checker,status IN ('active','matched','cancelled','expired')" json:"status"`
	MatchedEventID *string                  `json:"matched_event_id,omitempty"`
	MatchedEvent   TriggerEvent             `gorm:"constraint:OnDelete:RESTRICT" json:"-"`
	ActivatedAt    time.Time                `gorm:"not null;<-:create" json:"activated_at"`
	MatchedAt      *time.Time               `json:"matched_at,omitempty"`
	NotifiedAt     *time.Time               `json:"notified_at,omitempty"`
	CancelledAt    *time.Time               `json:"cancelled_at,omitempty"`
	ExpiredAt      *time.Time               `json:"expired_at,omitempty"`
}

func (w *EventRunbookWaiter) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, w, "active_trigger"), Columns: []string{"org_id", "trigger_id"}, Option: "WHERE status = 'active'"},
		{Name: indexes.Name(db, w, "unnotified"), Columns: []string{"matched_event_id"}, Option: "WHERE status = 'matched' AND notified_at IS NULL"},
	}
}

func (w *EventRunbookWaiter) BeforeCreate(*gorm.DB) error {
	if w.ID == "" {
		w.ID = domains.NewEventRunbookWaiterID()
	}
	return nil
}
