package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// DatadogManagedMonitorTargetType identifies which Nuon entity a managed
// monitor watches. Drives the query template used when creating the
// monitor in DD.
type DatadogManagedMonitorTargetType string

const (
	DatadogManagedMonitorTargetTypeAction    DatadogManagedMonitorTargetType = "action"
	DatadogManagedMonitorTargetTypeInstall   DatadogManagedMonitorTargetType = "install"
	DatadogManagedMonitorTargetTypeComponent DatadogManagedMonitorTargetType = "component"
	DatadogManagedMonitorTargetTypeWorkflow  DatadogManagedMonitorTargetType = "workflow"
)

// DatadogManagedMonitorPreset names the query template applied when
// creating the monitor. v1 ships two presets — see the DD client's
// monitor template package for the actual queries.
type DatadogManagedMonitorPreset string

const (
	DatadogManagedMonitorPresetFailure DatadogManagedMonitorPreset = "failure"
	DatadogManagedMonitorPresetDrift   DatadogManagedMonitorPreset = "drift"
)

// DatadogManagedMonitorStatus tracks whether a managed monitor is
// currently emitting alerts. Soft-delete still applies for hard deletes;
// "disabled" is the user-facing "pause this monitor without losing the
// row" state.
type DatadogManagedMonitorStatus string

const (
	DatadogManagedMonitorStatusActive   DatadogManagedMonitorStatus = "active"
	DatadogManagedMonitorStatusDisabled DatadogManagedMonitorStatus = "disabled"
)

// DatadogManagedMonitor records a DD monitor created by Nuon on behalf of
// the user via the one-click "Alert on failure in Datadog" button. The
// (connection_id, target_type, target_id, preset) tuple is unique so the
// button is idempotent — clicking twice never produces a duplicate.
//
// DDMonitorID is DD's int64 monitor ID returned from the Monitors API.
// Stored as int64 so we can hit DD's update/delete endpoints without
// extra parsing.
type DatadogManagedMonitor struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	ConnectionID string            `json:"connection_id,omitzero" gorm:"notnull;index:idx_datadog_managed_monitors_conn;uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"connection_id,omitzero,omitempty"`
	Connection   DatadogConnection `json:"-" gorm:"foreignKey:ConnectionID;references:ID;constraint:OnDelete:CASCADE" temporaljson:"connection,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_datadog_managed_monitors_org" temporaljson:"org_id,omitzero,omitempty"`

	TargetType DatadogManagedMonitorTargetType `json:"target_type,omitzero" gorm:"notnull;uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"target_type,omitzero,omitempty"`
	TargetID   string                          `json:"target_id,omitzero" gorm:"notnull;index:idx_datadog_managed_monitors_target;uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"target_id,omitzero,omitempty"`

	Preset DatadogManagedMonitorPreset `json:"preset,omitzero" gorm:"notnull;uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"preset,omitzero,omitempty"`

	DDMonitorID int64 `json:"dd_monitor_id,omitzero" gorm:"notnull" temporaljson:"dd_monitor_id,omitzero,omitempty"`

	Status DatadogManagedMonitorStatus `json:"status,omitzero" gorm:"notnull;default:'active'" temporaljson:"status,omitzero,omitempty"`

	// NotifyHandles snapshots the DD @-handles spliced into the monitor
	// body at creation time. Mirrors what's actually in DD so the
	// dashboard can show "alerting @pagerduty-prod" without round-tripping
	// to DD.
	NotifyHandles pq.StringArray `json:"notify_handles,omitzero" gorm:"type:text[]" swaggertype:"array,string" temporaljson:"notify_handles,omitzero,omitempty"`
}

func (DatadogManagedMonitor) TableName() string {
	return "datadog_managed_monitors"
}

func (a *DatadogManagedMonitor) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewDatadogManagedMonitorID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.Status == "" {
		a.Status = DatadogManagedMonitorStatusActive
	}

	return nil
}
