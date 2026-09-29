package labels

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// Labels defines a custom type for map[string]string that works with JSONB in GORM.
type Labels map[string]string

type Labeled struct {
	Labels Labels `json:"labels,omitzero" gorm:"default null" temporaljson:"labels,omitzero,omitempty"`
}

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

func (l Labels) Value() (driver.Value, error) {
	if l == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(l)
}

func (l Labels) GormDataType() string {
	return "jsonb"
}

func (l Labels) HasLabel(key, value string) bool {
	v, ok := l[key]
	return ok && v == value
}

func (l *Labels) Merge(other Labels) {
	if *l == nil {
		*l = make(Labels)
	}
	for k, v := range other {
		(*l)[k] = v
	}
}

func (l *Labels) RemoveKeys(keys []string) {
	if *l == nil {
		return
	}
	for _, k := range keys {
		delete(*l, k)
	}
}

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

		key, value, ok := strings.Cut(part, ":")
		if !ok {
			key, value, ok = strings.Cut(part, "=")
		}
		if !ok {
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
