package app

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// Lock a lock on state
type TerraformLock struct {
	Created   string `temporaljson:"created,omitzero,omitempty"`
	Path      string `temporaljson:"path,omitzero,omitempty"`
	ID        string `temporaljson:"id,omitzero,omitempty"`
	Operation string `temporaljson:"operation,omitzero,omitempty"`
	Info      string `temporaljson:"info,omitzero,omitempty"`
	Who       string `temporaljson:"who,omitzero,omitempty"`
	Version   any    `temporaljson:"version,omitzero,omitempty"`
}

func (c *TerraformLock) Scan(v interface{}) (err error) {
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
func (c *TerraformLock) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (TerraformLock) GormDataType() string {
	return "jsonb"
}
