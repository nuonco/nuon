package app

import (
	"time"
)

// InstallComponentHealthTransition is a ClickHouse row recorded by the
// component-health evaluator whenever a component's debounced health verdict
// changes. Powers the health timeline and uptime math; diagnosis and deploy
// correlation are populated in later phases.
type InstallComponentHealthTransition struct {
	OrgID              string `gorm:"column:org_id;type:LowCardinality(String)" json:"org_id"`
	InstallID          string `gorm:"column:install_id"                         json:"install_id"`
	InstallComponentID string `gorm:"column:install_component_id"               json:"install_component_id"`
	ComponentID        string `gorm:"column:component_id;default:''"            json:"component_id"`

	FromHealth string `gorm:"column:from_health;type:LowCardinality(String)" json:"from_health"`
	ToHealth   string `gorm:"column:to_health;type:LowCardinality(String)"   json:"to_health"`

	RootResourceKind      string `gorm:"column:root_resource_kind;default:''"      json:"root_resource_kind"`
	RootResourceNamespace string `gorm:"column:root_resource_namespace;default:''" json:"root_resource_namespace"`
	RootResourceName      string `gorm:"column:root_resource_name;default:''"      json:"root_resource_name"`
	Message               string `gorm:"column:message;default:''"                 json:"message"`

	Diagnosis          string `gorm:"column:diagnosis;default:''"            json:"diagnosis"`
	CorrelatedDeployID string `gorm:"column:correlated_deploy_id;default:''" json:"correlated_deploy_id"`

	ObservedAt time.Time `gorm:"column:observed_at;type:DateTime64(9);codec:Delta(8),ZSTD(1)" json:"observed_at"`
}

func (InstallComponentHealthTransition) TableName() string {
	return "install_component_health_transitions"
}

func (InstallComponentHealthTransition) GetTableOptions() string {
	return `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/install_component_health_transitions', '{replica}')
	TTL toDateTime(observed_at) + toIntervalDay(90)
	PARTITION BY toDate(observed_at)
	PRIMARY KEY (install_id, install_component_id, observed_at)
	ORDER BY    (install_id, install_component_id, observed_at)
	SETTINGS index_granularity = 8192`
}

func (InstallComponentHealthTransition) GetTableClusterOptions() string {
	return "on cluster simple"
}
