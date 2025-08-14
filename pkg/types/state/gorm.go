package state

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// Scan implements the database/sql.Scanner interface.
func (c *State) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan composite status")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c *State) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (State) GormDataType() string {
	return "jsonb"
}
