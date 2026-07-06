package compositeerrors

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SchemaVersion is the current version of the frozen CompositeErrorData
// payload. It is stamped onto every record by New() so renderers and future
// migrations can reason about payloads written by older code.
const SchemaVersion = 1

// CompositeErrorData is the JSONB GORM column attached to owner rows. It
// captures a typed CompositeError's payload along with its headline message
// and structured sections, all frozen at write time.
//
// To use, add this to a GORM model struct:
//
//	CompositeError *compositeerrors.CompositeErrorData `json:"composite_error,omitempty" gorm:"type:jsonb"`
type CompositeErrorData struct {
	// Version is the payload schema version (SchemaVersion at write time).
	Version int `json:"version"`

	Type     Type      `json:"type"`
	Severity Severity  `json:"severity"`
	Message  string    `json:"message"`
	Sections []Section `json:"sections,omitempty"`

	// Data is the typed, per-error-type payload: WHAT the error is. Closed
	// schema per Type. Read to render sections and by any future view.
	Data json.RawMessage `json:"data" swaggertype:"object"`

	// SourceID / SourceType identify the row this error originated on
	// (polymorphic, same shape as OwnerID/OwnerType). Set at the record site,
	// e.g. ("runner_job_execution_results", "<result id>"). Enables a future
	// JOINable view without a separate error table.
	SourceID   string `json:"source_id,omitempty"`
	SourceType string `json:"source_type,omitempty"`

	// Hints is the open annotation/directive bag: HOW to handle or present the
	// error. Canonical keys (Hint*) are honored by specific consumers.
	Hints Hints `json:"hints,omitempty"`
}

// Option customizes a CompositeErrorData at construction. Options apply
// record-site context (e.g. provenance) that the typed error doesn't know.
type Option func(*CompositeErrorData)

// WithSource records where the error originated. sourceType is the row kind
// (polymorphic table name, e.g. "runner_job_execution_results"); sourceID is
// that row's id.
func WithSource(sourceType, sourceID string) Option {
	return func(d *CompositeErrorData) {
		d.SourceType = sourceType
		d.SourceID = sourceID
	}
}

// New constructs a CompositeErrorData from a typed CompositeError. The
// implementation's data, headline message, sections, and (optionally) hints
// are captured at this point and frozen on the resulting record. Options apply
// record-site context such as provenance.
func New(e CompositeError, opts ...Option) *CompositeErrorData {
	data, _ := json.Marshal(e)
	d := &CompositeErrorData{
		Version:  SchemaVersion,
		Type:     e.Type(),
		Severity: e.Severity(),
		Message:  e.Error(),
		Sections: e.Sections(),
		Data:     data,
	}
	if hp, ok := e.(HintsProvider); ok {
		if hints := hp.Hints(); len(hints) > 0 {
			d.Hints = hints
		}
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Scan implements database/sql.Scanner.
func (c *CompositeErrorData) Scan(value any) error {
	if value == nil {
		*c = CompositeErrorData{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("compositeerrors: cannot scan type %T", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*c = CompositeErrorData{}
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer.
func (c *CompositeErrorData) Value() (driver.Value, error) {
	if c == nil || c.Type == "" {
		return nil, nil
	}
	return json.Marshal(c)
}

// GormDataType tells GORM to use a jsonb column.
func (CompositeErrorData) GormDataType() string { return "jsonb" }
