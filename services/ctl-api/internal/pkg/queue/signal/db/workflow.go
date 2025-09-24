package signaldb

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type WorkflowRef struct {
	IDTemplate string `json:"-"`

	Namespace string `json:"namespace"`
	ID        string `json:"id"`
}

func (s WorkflowRef) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *WorkflowRef) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type")
	}

	if err := json.Unmarshal(bytes, s); err != nil {
		return errors.Wrap(err, "unable to convert workflow json to ref")
	}

	return nil
}

func (WorkflowRef) GormDataType() string {
	return "jsonb"
}
