package compositeerrors

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
)

// SchemaVersion is the current version of the frozen CompositeErrorData
// payload. It is stamped onto every record by New() so renderers and future
// migrations can reason about payloads written by older code.
const SchemaVersion = 1

const redactedValue = "[REDACTED]"

const sensitiveCredentialName = `(?:access[-_]?key|access[-_]?token|api[-_]?key|auth[-_]?token|awsaccesskeyid|client[-_]?secret|id[-_]?token|password|refresh[-_]?token|secret|session[-_]?token|sig|signature|token|x-amz-credential|x-amz-security-token|x-amz-signature|x-goog-credential|x-goog-signature)`
const diagnosticCredentialName = `(?:[[:alnum:]]+[-_.])*` + sensitiveCredentialName

type diagnosticSecretRedactor struct {
	pattern     *regexp.Regexp
	replacement string
}

var diagnosticSecretRedactors = []diagnosticSecretRedactor{
	{
		// Credentials embedded in URL userinfo, such as
		// https://user:password@host or postgres://user:password@host.
		pattern:     regexp.MustCompile(`(?i)(\b(?:https?|ssh|git|postgres(?:ql)?|mysql|redis|mongodb(?:\+srv)?|mssql|sqlserver|amqps?)://)[^/@\s]+@`),
		replacement: "${1}" + redactedValue + "@",
	},
	{
		// HTTP authorization and cookie header values.
		pattern:     regexp.MustCompile(`(?im)(\b(?:authorization|proxy-authorization|cookie|set-cookie)[ \t]*:[ \t]*)[^\r\n]*`),
		replacement: "${1}" + redactedValue,
	},
	{
		// Quoted JSON-style credential fields, such as "api_token":"...".
		pattern:     regexp.MustCompile(`(?i)(["']` + diagnosticCredentialName + `["']\s*:\s*["'])[^"'\r\n]*`),
		replacement: "${1}" + redactedValue,
	},
	{
		// Quoted HCL, shell, or environment assignments, such as password = "...".
		pattern:     regexp.MustCompile(`(?i)(\b` + diagnosticCredentialName + `\s*=\s*["'])[^"'\r\n]*`),
		replacement: "${1}" + redactedValue,
	},
	{
		// Unquoted URL query and environment assignments, such as token=<value>.
		pattern:     regexp.MustCompile(`(?i)(\b` + diagnosticCredentialName + `\s*=\s*)[^&\s"'<>]+`),
		replacement: "${1}" + redactedValue,
	},
}

// Credential fields encountered after typed composite-error data is decoded.
var sensitiveJSONKey = regexp.MustCompile(`(?i)^` + diagnosticCredentialName + `$`)

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
// are captured at this point and frozen on the resulting record. Sections and
// Hints are copied so a shared source (e.g. a package-level default hints bag)
// cannot be mutated through the persisted record. Options apply record-site
// context such as provenance.
//
// It returns an error when the typed payload cannot be marshalled, so a
// serialization failure surfaces at the call site instead of silently
// persisting a record with a null payload.
func New(e CompositeError, opts ...Option) (*CompositeErrorData, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("compositeerrors: marshal %T payload: %w", e, err)
	}
	d := &CompositeErrorData{
		Version:  SchemaVersion,
		Type:     e.Type(),
		Severity: e.Severity(),
		Message:  e.Error(),
		Sections: cloneSections(e.Sections()),
		Data:     data,
	}
	if hp, ok := e.(HintsProvider); ok {
		d.Hints = hp.Hints().Clone()
	}
	for _, opt := range opts {
		opt(d)
	}
	redacted, err := d.redacted()
	if err != nil {
		return nil, fmt.Errorf("compositeerrors: redact %T payload: %w", e, err)
	}
	return &redacted, nil
}

// cloneSections returns a copy of s, or nil when empty. Section fields are all
// value types, so a slice copy fully detaches the result from the source.
func cloneSections(s []Section) []Section {
	if len(s) == 0 {
		return nil
	}
	out := make([]Section, len(s))
	copy(out, s)
	return out
}

func (c CompositeErrorData) redacted() (CompositeErrorData, error) {
	c.Message = RedactDiagnosticSecrets(c.Message)
	c.Sections = cloneSections(c.Sections)
	for i := range c.Sections {
		c.Sections[i].Body = RedactDiagnosticSecrets(c.Sections[i].Body)
	}
	c.Hints = c.Hints.Clone()
	for key, value := range c.Hints {
		c.Hints[key] = RedactDiagnosticSecrets(value)
	}

	data, err := redactJSONStrings(c.Data)
	if err != nil {
		return CompositeErrorData{}, err
	}
	c.Data = data
	return c, nil
}

// RedactDiagnosticSecrets removes credentials in common URL, header, and
// structured-assignment forms. It is intentionally not a general-purpose
// detector for arbitrary secrets embedded in otherwise unlabeled prose.
func RedactDiagnosticSecrets(value string) string {
	for _, redactor := range diagnosticSecretRedactors {
		value = redactor.pattern.ReplaceAllString(value, redactor.replacement)
	}
	return value
}

func redactJSONStrings(data json.RawMessage) (json.RawMessage, error) {
	if len(data) == 0 {
		return data, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	value, changed := redactJSONValue(value)
	if !changed {
		return data, nil
	}
	return json.Marshal(value)
}

func redactJSONValue(value any) (any, bool) {
	switch value := value.(type) {
	case string:
		redacted := RedactDiagnosticSecrets(value)
		return redacted, redacted != value
	case []any:
		var changed bool
		for i := range value {
			var itemChanged bool
			value[i], itemChanged = redactJSONValue(value[i])
			changed = changed || itemChanged
		}
		return value, changed
	case map[string]any:
		var changed bool
		for key := range value {
			if _, ok := value[key].(string); ok && sensitiveJSONKey.MatchString(key) {
				value[key] = redactedValue
				changed = true
				continue
			}
			var itemChanged bool
			value[key], itemChanged = redactJSONValue(value[key])
			changed = changed || itemChanged
		}
		return value, changed
	}
	return value, false
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
