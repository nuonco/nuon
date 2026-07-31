package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// DLQRecord is a record a consumer could not decode: the envelope itself
// failed to parse, or its version is newer than this binary understands, or
// its payload didn't unmarshal into the expected type. It is NOT for insert
// failures (ClickHouse down, timeout) — those stay on the normal redelivery
// path, since the record itself is fine and just needs a retry.
//
// One table for every consumer; Topic/ConsumerName disambiguate origin. See
// plans/08-kafka-phase5-consumer-hardening.md for the fuller rationale.
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

	// EnvelopeType is best-effort: empty when the envelope itself didn't parse.
	EnvelopeType string    `json:"envelope_type,omitzero" temporaljson:"envelope_type,omitzero,omitempty"`
	ProducedAt   time.Time `json:"produced_at,omitzero" temporaljson:"produced_at,omitzero,omitempty"`
	FailedAt     time.Time `json:"failed_at,omitzero" temporaljson:"failed_at,omitzero,omitempty"`

	// RawValue is the undecoded Kafka record value, stored as-is rather than
	// base64-encoded. This value is often exactly the malformed data Unwrap
	// rejected, so it isn't guaranteed to be valid UTF-8 — encoding/json
	// replaces invalid UTF-8 with U+FFFD on marshal, which does lose those
	// bytes. Deliberate: readability wins here. A dead letter is inspected by
	// a human reading it directly in ClickHouse most of the time, and a
	// visible run of U+FFFD is itself a useful signal — it marks exactly
	// where the record broke — versus a base64 blob that has to be decoded
	// out-of-band before any of the surrounding, still-legible message can be
	// read at all.
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
