package app

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// Lock a lock on state
type TerraformLock struct {
	Created   string
	Path      string
	ID        string
	Operation string
	Info      string
	Who       string
	Version   any
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
