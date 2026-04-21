package app

import (
	"time"
)

// PolicyReportEvent is a ClickHouse analytics event for policy evaluations.
// One row per policy per report, with outcome derived from PolicyResult.Status.
type PolicyReportEvent struct {
	ReportID    string    `gorm:"column:report_id"                               json:"report_id"`
	OrgID       string    `gorm:"column:org_id;type:LowCardinality(String)"      json:"org_id"`
	AppID       string    `gorm:"column:app_id"                                  json:"app_id"`
	InstallID   string    `gorm:"column:install_id;default:''"                   json:"install_id"`
	ComponentID string    `gorm:"column:component_id;default:''"                 json:"component_id"`
	PolicyID    string    `gorm:"column:policy_id"                               json:"policy_id"`
	OwnerType   string    `gorm:"column:owner_type;type:LowCardinality(String)"  json:"owner_type"`
	EvaluatedAt time.Time `gorm:"column:evaluated_at;type:DateTime64(3)"         json:"evaluated_at"`
	Outcome     string    `gorm:"column:outcome;type:LowCardinality(String)"     json:"outcome"`
}

func (PolicyReportEvent) TableName() string {
	return "policy_report_events"
}

func (PolicyReportEvent) GetTableOptions() string {
	return `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/policy_report_events', '{replica}')
	TTL toDateTime(evaluated_at) + toIntervalDay(180)
	PARTITION BY toYYYYMM(evaluated_at)
	PRIMARY KEY (org_id, app_id, evaluated_at, policy_id)
	ORDER BY    (org_id, app_id, evaluated_at, policy_id)
	SETTINGS index_granularity = 8192`
}

func (PolicyReportEvent) GetTableClusterOptions() string {
	return "on cluster simple"
}
