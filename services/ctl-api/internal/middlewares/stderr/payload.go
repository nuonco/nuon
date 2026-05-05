package stderr

import (
	"errors"
)

// StepErrorPayload is the structured side-channel attached to a Temporal
// NonRetryableApplicationError so the workflow side can recover the typed
// ErrUser fields after the activity → workflow boundary strips Go error
// types.
//
// It is also serialized into the QueueSignal status Metadata under the
// MetadataKey* keys, which is the durable source of truth; the payload is
// the in-flight transport copy.
//
// Keep this struct flat and JSON-encodable (the default Temporal data
// converter is JSON).
type StepErrorPayload struct {
	Code      string            `json:"code,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Directive string            `json:"directive,omitempty"`
}

// IsZero reports whether the payload carries no useful data and can be
// omitted from a Temporal ApplicationError.
func (p StepErrorPayload) IsZero() bool {
	return p.Code == "" && p.Directive == "" && len(p.Fields) == 0
}

// Metadata keys used to persist StepErrorPayload fields on a QueueSignal
// status Metadata map. Read by AwaitSignal when reconstructing the
// payload from the DB.
const (
	MetadataKeyErrorCode     = "error_code"
	MetadataKeyErrorFields   = "error_fields"
	MetadataKeyStepDirective = "step_directive"
)

// PayloadFromErrUser builds a StepErrorPayload from an ErrUser. The two
// types share the same shape on the wire — this is the canonical converter.
func PayloadFromErrUser(u ErrUser) StepErrorPayload {
	return StepErrorPayload{
		Code:      u.Code,
		Fields:    u.Fields,
		Directive: string(u.Directive),
	}
}

// MetadataFromPayload converts a StepErrorPayload to a metadata map
// suitable for merging into a QueueSignal status Metadata. Empty fields
// are omitted so the metadata map stays minimal. Returns nil when the
// payload is zero.
func MetadataFromPayload(p StepErrorPayload) map[string]any {
	if p.IsZero() {
		return nil
	}
	meta := map[string]any{}
	if p.Code != "" {
		meta[MetadataKeyErrorCode] = p.Code
	}
	if len(p.Fields) > 0 {
		meta[MetadataKeyErrorFields] = p.Fields
	}
	if p.Directive != "" {
		meta[MetadataKeyStepDirective] = p.Directive
	}
	return meta
}

// MetadataFromErrUser is a convenience wrapper combining PayloadFromErrUser
// and MetadataFromPayload. Returns nil when the ErrUser has no
// payload-relevant fields.
func MetadataFromErrUser(u ErrUser) map[string]any {
	return MetadataFromPayload(PayloadFromErrUser(u))
}

// PayloadFromMeta extracts a StepErrorPayload from a QueueSignal status
// Metadata map, tolerating the JSON round-trip — when metadata comes from
// the DB JSONB column, nested values come back as map[string]any.
func PayloadFromMeta(meta map[string]any) StepErrorPayload {
	var p StepErrorPayload
	if meta == nil {
		return p
	}
	if v, ok := meta[MetadataKeyErrorCode].(string); ok {
		p.Code = v
	}
	if v, ok := meta[MetadataKeyStepDirective].(string); ok {
		p.Directive = v
	}
	p.Fields = coerceStringMap(meta[MetadataKeyErrorFields])
	return p
}

// ErrUserFromPayload reconstructs an ErrUser from a StepErrorPayload.
// msg is used as both the underlying error message and the user-facing
// description.
func ErrUserFromPayload(p StepErrorPayload, msg string) ErrUser {
	return ErrUser{
		Err:         errors.New(msg),
		Description: msg,
		Code:        p.Code,
		Fields:      p.Fields,
		Directive:   StepDirective(p.Directive),
	}
}

// ErrUserFromMeta is a convenience wrapper combining PayloadFromMeta and
// ErrUserFromPayload. Returns the zero value and false when meta has no
// payload-relevant fields.
func ErrUserFromMeta(meta map[string]any, msg string) (ErrUser, bool) {
	p := PayloadFromMeta(meta)
	if p.IsZero() {
		return ErrUser{}, false
	}
	return ErrUserFromPayload(p, msg), true
}

// coerceStringMap converts a map-typed any into a map[string]string,
// handling both the direct map[string]string case and the JSON-decoded
// map[string]any case.
func coerceStringMap(v any) map[string]string {
	switch f := v.(type) {
	case map[string]string:
		return f
	case map[string]any:
		out := make(map[string]string, len(f))
		for k, vv := range f {
			if s, ok := vv.(string); ok {
				out[k] = s
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return nil
}
