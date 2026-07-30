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

func (i *InstallComponentResourceState) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &InstallComponentResourceState{}, 1),
			SQL:           viewsql.InstallComponentResourceStatesViewV1,
			AlwaysReapply: true,
		},
	}
}
