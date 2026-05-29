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

// DatadogManagedMonitorMode selects how the monitor evaluates Nuon
// lifecycle outcomes. There are two modes; both fire the same DD monitor
// alert on the user's side, but they differ in where the matching
// happens:
//
//   - event: DD's event-v2 query language matches over the DD event
//     stream. Requires the org to have a verified DD event subscription
//     because the events have to be flowing into DD for DD to query
//     them. Original v1 mode; default for back-compat.
//
//   - metric: Nuon evaluates match / interests on its own side via the
//     signal lifecycle hook and submits a single low-cardinality metric
//     `nuon.monitor.fired{nuon_monitor_id:<id>}` to DD. The DD monitor
//     is a sum(last_5m) > 0 metric alert on that one tag value.
//     Decouples DD-side alerting from the event-subscription path —
//     useful when an org wants alerts in DD without routing every Nuon
//     event through the DD event stream.
//
// Metric mode keeps DD custom-metric cardinality bounded: only one tag
// (nuon_monitor_id) is ever submitted, so the total series count equals
// the number of managed-monitor rows regardless of how many installs /
// components / actions / labels feed into the matcher. Install / action
// / label filtering happens inside the Nuon hook before the metric is
// ever submitted.
type DatadogManagedMonitorMode string

const (
	DatadogManagedMonitorModeEvent  DatadogManagedMonitorMode = "event"
	DatadogManagedMonitorModeMetric DatadogManagedMonitorMode = "metric"
)

// DatadogManagedMonitor records a DD monitor created by Nuon on behalf of
// the user via the one-click "Alert on failure in Datadog" button. The
// (connection_id, target_type, target_id, install_id, preset) tuple is
// unique so the button is idempotent — clicking twice on the same
// (target, install) scope never produces a duplicate. InstallID is the
// optional install-scope used by action-target monitors (org-level
// action_workflows.id needs an install qualifier to mirror per-install
// targeting); empty for install/component/workflow which are already
// install-scoped by their own targetID.
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

	// InstallID is set on action-target monitors to scope the alert to
	// one install's invocations of that action. Empty for the other
	// target types (their TargetID already carries install scope) and
	// for org-wide action monitors if we ever support those.
	InstallID string `json:"install_id,omitzero" gorm:"default:'';uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"install_id,omitzero,omitempty"`

	Preset DatadogManagedMonitorPreset `json:"preset,omitzero" gorm:"notnull;uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"preset,omitzero,omitempty"`

	DDMonitorID int64 `json:"dd_monitor_id,omitzero" gorm:"notnull" temporaljson:"dd_monitor_id,omitzero,omitempty"`

	Status DatadogManagedMonitorStatus `json:"status,omitzero" gorm:"notnull;default:'active'" temporaljson:"status,omitzero,omitempty"`

	// Mode selects event-stream vs metric-driven matching. Defaults to
	// "event" via the DB-level default so existing v1 rows keep their
	// original behavior without a backfill. See the type doc for the
	// tradeoff between modes.
	//
	// Included in the unique index so the same (connection, target,
	// preset) tuple can host one event-mode AND one metric-mode monitor
	// — e.g. during a migration where the user is moving alerts off the
	// event-subscription path onto metric mode but doesn't want a gap.
	Mode DatadogManagedMonitorMode `json:"mode,omitzero" gorm:"notnull;default:'event';uniqueIndex:idx_datadog_managed_monitors_conn_target_preset" temporaljson:"mode,omitzero,omitempty"`

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

	if a.Mode == "" {
		a.Mode = DatadogManagedMonitorModeEvent
	}

	return nil
}
