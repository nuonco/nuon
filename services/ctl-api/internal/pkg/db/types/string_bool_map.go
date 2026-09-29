package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringBoolMap defines a custom type for map[string]bool that works with JSONB
type StringBoolMap map[string]bool

func (m *StringBoolMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(StringBoolMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for StringBoolMap")
	}

	if err := json.Unmarshal(bytes, m); err != nil {
		return err
	}

	return nil
}

func (m StringBoolMap) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]bool{})
	}
	return json.Marshal(m)
}
