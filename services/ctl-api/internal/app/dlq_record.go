package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type DLQRecord struct {
	ID        string    `gorm:"primary_key" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`

	Topic         string `json:"topic,omitzero" temporaljson:"topic,omitzero,omitempty"`
	Partition     int32  `json:"partition,omitzero" temporaljson:"partition,omitzero,omitempty"`
	Offset        int64  `json:"offset,omitzero" temporaljson:"offset,omitzero,omitempty"`
	ConsumerGroup string `json:"consumer_group,omitzero" temporaljson:"consumer_group,omitzero,omitempty"`
	ConsumerName  string `json:"consumer_name,omitzero" temporaljson:"consumer_name,omitzero,omitempty"`

	Reason string `json:"reason,omitzero" temporaljson:"reason,omitzero,omitempty"`
	Error  string `json:"error,omitzero" temporaljson:"error,omitzero,omitempty"`

	EnvelopeType string    `json:"envelope_type,omitzero" temporaljson:"envelope_type,omitzero,omitempty"`
	ProducedAt   time.Time `json:"produced_at,omitzero" temporaljson:"produced_at,omitzero,omitempty"`
	FailedAt     time.Time `json:"failed_at,omitzero" temporaljson:"failed_at,omitzero,omitempty"`

	RawValue string `json:"raw_value,omitzero" temporaljson:"raw_value,omitzero,omitempty"`
}

func (r *DLQRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewDLQRecordID()
	}
	if r.FailedAt.IsZero() {
		r.FailedAt = time.Now()
	}
	return nil
}

func (r DLQRecord) TableName() string {
	return "dlq"
}

func (r DLQRecord) GetTableOptions() string {
	return `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/dlq', '{replica}')
	TTL toDateTime(failed_at) + toIntervalDay(30)
	PARTITION BY toDate(failed_at)
	ORDER BY    (topic, failed_at)`
}

func (r DLQRecord) GetTableClusterOptions() string {
	return "on cluster simple"
}
