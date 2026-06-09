package lifecyclephase

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Phase string

const (
	Provisioning   Phase = "provisioning"
	Provisioned    Phase = "provisioned"
	Deprovisioning Phase = "deprovisioning"
	Deprovisioned  Phase = "deprovisioned"
	Reprovisioning Phase = "reprovisioning"
)

type LifecyclePhase struct {
	Phase       Phase            `json:"phase,omitempty" temporaljson:"phase,omitempty"`
	Description string           `json:"description,omitempty" temporaljson:"description,omitempty"`
	UpdatedAt   int64            `json:"updated_at,omitempty" temporaljson:"updated_at,omitempty"`
	UpdatedBy   string           `json:"updated_by,omitempty" temporaljson:"updated_by,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty" temporaljson:"metadata,omitempty"`
	History     []LifecyclePhase `json:"history,omitempty" temporaljson:"history,omitempty"`
}

func New(phase Phase, description string) LifecyclePhase {
	return LifecyclePhase{
		Phase:       phase,
		Description: description,
		UpdatedAt:   time.Now().Unix(),
		Metadata:    make(map[string]any),
	}
}

func Transition(current LifecyclePhase, next Phase, description string) LifecyclePhase {
	prev := current
	prev.History = nil

	history := append([]LifecyclePhase{prev}, current.History...)
	if len(history) > 25 {
		history = history[:25]
	}

	return LifecyclePhase{
		Phase:       next,
		Description: description,
		UpdatedAt:   time.Now().Unix(),
		UpdatedBy:   current.UpdatedBy,
		Metadata:    current.Metadata,
		History:     history,
	}
}

func (lp *LifecyclePhase) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for LifecyclePhase")
	}

	return json.Unmarshal(bytes, lp)
}

func (lp LifecyclePhase) Value() (driver.Value, error) {
	return json.Marshal(lp)
}

func (lp LifecyclePhase) GormDataType() string {
	return "jsonb"
}
