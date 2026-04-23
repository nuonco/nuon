package signaldb

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type UpdateHandler struct {
	Namespace  string `json:"namespace"`
	WorkflowID string `json:"workflow_id"`
	UpdateName string `json:"update_name"`
}

func (u UpdateHandler) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *UpdateHandler) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type")
	}

	if err := json.Unmarshal(bytes, u); err != nil {
		return errors.Wrap(err, "unable to convert update handler json to ref")
	}

	return nil
}

func (UpdateHandler) GormDataType() string {
	return "jsonb"
}
