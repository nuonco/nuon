package app

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type Signal struct {
	Namespace  string `json:"namespace,omitzero" temporaljson:"namespace,omitzero,omitempty"`
	Type       string `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	SignalJSON []byte `json:"json,omitzero" temporaljson:"signal_json,omitzero,omitempty"`
}

func (s *Signal) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if string(v) == "null" {
			return nil
		}
		if err := json.Unmarshal(v, s); err != nil {
			return errors.Wrap(err, "unable to scan composite status")
		}
	}

	return
}

func (s *Signal) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (Signal) GormDataType() string {
	return "jsonb"
}
