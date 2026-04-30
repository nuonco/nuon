package cctx

import (
	"database/sql/driver"
	"encoding/json"
)

// SignalContext captures the request-scoped context values that must survive
// asynchronous enqueuing. It is stored as a JSONB column on QueueSignal so that
// the background enqueuer can restore the full context when processing the
// signal outside the original request goroutine.
type SignalContext struct {
	AccountID   string `json:"account_id,omitempty"`
	OrgID       string `json:"org_id,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}

// Scan implements database/sql.Scanner for reading JSONB from PostgreSQL.
func (sc *SignalContext) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		return json.Unmarshal(v, sc)
	}
	return nil
}

// Value implements driver.Valuer for writing JSONB to PostgreSQL.
func (sc SignalContext) Value() (driver.Value, error) {
	return json.Marshal(sc)
}

// GormDataType tells GORM to use the jsonb PostgreSQL type.
func (SignalContext) GormDataType() string {
	return "jsonb"
}
