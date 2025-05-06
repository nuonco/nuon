package app

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// Lock a lock on state
type TerraformLock struct {
	Created   string `json:"created,omitzero" temporaljson:"created,omitzero,omitempty"`
	Path      string `json:"path,omitzero" temporaljson:"path,omitzero,omitempty"`
	ID        string `json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	Operation string `json:"operation,omitzero" temporaljson:"operation,omitzero,omitempty"`
	Info      string `json:"info,omitzero" temporaljson:"info,omitzero,omitempty"`
	Who       string `json:"who,omitzero" temporaljson:"who,omitzero,omitempty"`
	Version   any    `json:"version,omitzero" temporaljson:"version,omitzero,omitempty"`
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
