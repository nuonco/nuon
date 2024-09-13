package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// Nested implements Scan and Value for the clickhouse Nested column type.
// Currently we are hijacking the JSON approach for this.

// Truly, we just need a type so we can accept this as a migrate-able field w/out errors.
// JSON is good here because we can then unmarshall the json and dynamically populate
// Our struct in AfterLoad

type Nested json.RawMessage

// Scan scans value into Jsonb, implements sql.Scanner interface
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

// Value return json value, implement driver.Valuer interface
func (n Nested) Value() (driver.Value, error) {
	if len(n) == 0 {
		return nil, nil
	}
	return json.RawMessage(n).MarshalJSON()
}
