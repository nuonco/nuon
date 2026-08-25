package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallComponentResourceHealth string

const (
	InstallComponentResourceHealthHealthy     InstallComponentResourceHealth = "healthy"
	InstallComponentResourceHealthProgressing InstallComponentResourceHealth = "progressing"
	InstallComponentResourceHealthDegraded    InstallComponentResourceHealth = "degraded"
	InstallComponentResourceHealthUnhealthy   InstallComponentResourceHealth = "unhealthy"
	InstallComponentResourceHealthUnknown     InstallComponentResourceHealth = "unknown"
)

const (
	InstallComponentResourceSourceComponent = "component"
	InstallComponentResourceSourceSandbox   = "sandbox"
)

// InstallComponentResourceStatesLatestView is the read path for latest state: a
// ReplacingMergeTree keyed by resource identity, wrapped in a FINAL view.
// Aggregating the observation table instead costs ~65 MB per read on prod to
// return a few dozen rows, because it re-reads every 60s snapshot in the TTL.
//
// Created by CH migration 09 rather than Views() below, since the views phase
// runs before custom migrations and so cannot reference the table it selects from.
const InstallComponentResourceStatesLatestView = "install_component_resource_states_latest_view_v1"

// InstallComponentResourceState is a ClickHouse observation row: one per
// resource per install component; latest-state view keeps the newest per identity.
type InstallComponentResourceState struct {
	OrgID              string `gorm:"column:org_id;type:LowCardinality(String)"      json:"org_id"`
	InstallID          string `gorm:"column:install_id"                              json:"install_id"`
	InstallComponentID string `gorm:"column:install_component_id"                    json:"install_component_id"`
	ComponentID        string `gorm:"column:component_id;default:''"                 json:"component_id"`
	RunnerID           string `gorm:"column:runner_id;default:''"                    json:"runner_id"`

	// Source classifies the resource owner: "component" (keyed by
	// install_component_id) or "sandbox" (keyed by owner_name = helm release name).
	Source    string `gorm:"column:source;type:LowCardinality(String);default:'component'" json:"source"`
	OwnerName string `gorm:"column:owner_name;default:''"                              json:"owner_name"`

	Provider  string `gorm:"column:provider;type:LowCardinality(String)"    json:"provider"`
	APIGroup  string `gorm:"column:api_group;default:''"                    json:"api_group"`
	Kind      string `gorm:"column:kind;default:''"                         json:"kind"`
	Namespace string `gorm:"column:namespace;default:''"                    json:"namespace"`
	Name      string `gorm:"column:name;default:''"                         json:"name"`

	Health       string `gorm:"column:health;type:LowCardinality(String)"      json:"health"`
	Message      string `gorm:"column:message;default:''"                      json:"message"`
	NativeStatus string `gorm:"column:native_status;default:''"                json:"native_status"`
	Details      string `gorm:"column:details;default:''"                      json:"details"`

	ObservedAt time.Time `gorm:"column:observed_at;type:DateTime64(9);codec:Delta(8),ZSTD(1)" json:"observed_at"`

	// StaleAfterSeconds is how long this observation stays trustworthy (0 =
	// default); a pushed check sets its own, since it knows its cadence best.
	StaleAfterSeconds uint32 `gorm:"column:stale_after_seconds;default:0" json:"stale_after_seconds,omitempty"`

	// RemovedFromConfig is set at read time when a probe's name is no longer in
	// the component's config — still shown, but labelled so it can't pass as live.
	RemovedFromConfig bool `gorm:"-" json:"removed_from_config,omitempty"`
}

func (InstallComponentResourceState) TableName() string {
	return "install_component_resource_states"
}

func (InstallComponentResourceState) GetTableOptions() string {
	return `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/install_component_resource_states', '{replica}')
	TTL toDateTime(observed_at) + toIntervalDay(7)
	PARTITION BY toDate(observed_at)
	PRIMARY KEY (install_id, install_component_id, kind, namespace, name, observed_at)
	ORDER BY    (install_id, install_component_id, kind, namespace, name, observed_at)
	SETTINGS index_granularity = 8192`
}

func (InstallComponentResourceState) GetTableClusterOptions() string {
	return "on cluster simple"
}

func (*InstallComponentResourceState) UseView() bool {
	return false
}

func (*InstallComponentResourceState) ViewVersion() string {
	return "v1"
}

// Views keeps the aggregating latest-state view alive even though nothing reads it
// any more — see InstallComponentResourceStatesLatestView for the live read path.
// Dropping it in the same release that repoints the readers would 5xx the pods still
// draining behind the rollout, on the exact endpoint this change exists to speed up.
// Safe to delete once a release carrying the constant above has fully rolled out.
func (i *InstallComponentResourceState) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &InstallComponentResourceState{}, 1),
			SQL:           viewsql.InstallComponentResourceStatesViewV1,
			AlwaysReapply: true,
		},
	}
}

// InstallComponentResourceProviderCustom marks a pushed check rather than an
// observation the runner made of the cluster.
const InstallComponentResourceProviderCustom = "custom"

// LatestReportOnlySQL restricts the latest-state view to resources a report
// group's most recent report still contained.
//
// Deletion is not representable in an append-only log: a removed resource stops
// being reported, and the view keeps its final row forever, so a deleted pod
// read degraded permanently. One report stamps every row it carries with a
// single observed_at, so anything behind a group's newest row is gone.
//
// Written as a predicate rather than applied to the result set because callers
// also filter on health, and a filtered set has no reliable newest row. Pushed
// checks are exempt: they arrive on their own cadence and expire by their TTL.
func LatestReportOnlySQL() string {
	return "(provider = ? OR (install_component_id, source, owner_name, observed_at) IN (" +
		"SELECT install_component_id, source, owner_name, max(observed_at) FROM " +
		InstallComponentResourceStatesLatestView +
		" WHERE org_id = ? AND install_id = ? AND provider != ?" +
		" GROUP BY install_component_id, source, owner_name))"
}

// LatestReportOnlyArgs are the bind values for LatestReportOnlySQL.
func LatestReportOnlyArgs(orgID, installID string) []any {
	return []any{
		InstallComponentResourceProviderCustom,
		orgID,
		installID,
		InstallComponentResourceProviderCustom,
	}
}
