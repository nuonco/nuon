package labels

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// Labels defines a custom type for map[string]string that works with JSONB in GORM.
type Labels map[string]string

// Labeled is an embeddable struct for any GORM model that supports labels.
// Embed it to get a consistent JSONB labels column with standard tags.
type Labeled struct {
	Labels Labels `json:"labels,omitzero" gorm:"default null" temporaljson:"labels,omitzero,omitempty"`
}

// Scan implements the sql.Scanner interface for database deserialization.
func (l *Labels) Scan(value interface{}) error {
	if value == nil {
		*l = make(Labels)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for Labels")
	}

	if err := json.Unmarshal(bytes, l); err != nil {
		return err
	}

	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (l Labels) Value() (driver.Value, error) {
	if l == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(l)
}

// GormDataType returns the GORM data type for this field.
func (l Labels) GormDataType() string {
	return "jsonb"
}

// HasLabel returns true if the label with the given key and value exists.
func (l Labels) HasLabel(key, value string) bool {
	v, ok := l[key]
	return ok && v == value
}

// Merge adds all key-value pairs from other into the receiver,
// overwriting existing keys.
func (l *Labels) Merge(other Labels) {
	if *l == nil {
		*l = make(Labels)
	}
	for k, v := range other {
		(*l)[k] = v
	}
}

// RemoveKeys removes the specified keys from the labels.
func (l *Labels) RemoveKeys(keys []string) {
	if *l == nil {
		return
	}
	for _, k := range keys {
		delete(*l, k)
	}
}

// ParseLabelsQuery parses a comma-separated "key:value" string into Labels.
// Returns nil for empty input. Entries without a separator are treated as
// wildcard key-only filters (value set to "*").
// Supports both ":" and "=" as key-value separators.
// Splits on the first separator only, so values may contain colons or equals.
// A value of "*" means "match any value for this key" (wildcard).
func ParseLabelsQuery(raw string) Labels {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	result := make(Labels)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try colon first, then equals sign.
		key, value, ok := strings.Cut(part, ":")
		if !ok {
			key, value, ok = strings.Cut(part, "=")
		}
		if !ok {
			// Bare key with no separator — treat as wildcard.
			key = part
			value = "*"
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" {
			result[key] = value
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
