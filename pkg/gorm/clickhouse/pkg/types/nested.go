package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type Nested json.RawMessage

func (n *Nested) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal Nested value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*n = Nested(result)
	return err
}

func (n Nested) Value() (driver.Value, error) {
	if len(n) == 0 {
		return nil, nil
	}
	return json.RawMessage(n).MarshalJSON()
}
