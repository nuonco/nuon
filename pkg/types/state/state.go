package state

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Org          *OrgState          `json:"org"`
	App          *AppState          `json:"app"`
	Sandbox      *SandboxState      `json:"sandbox"`
	Inputs       *InputsState       `json:"inputs"`
	Actions      *ActionsState      `json:"actions"`
	Runner       *RunnerState       `json:"runner"`
	Components   map[string]any     `json:"components"`
	Domain       *DomainState       `json:"domain"`
	Cloud        *CloudAccount      `json:"cloud_account"`
	InstallStack *InstallStackState `json:"install_stack"`
	Secrets      *SecretsState      `json:"secrets"`

	// NOTE: for backwards compatibility, these are remaining in place.
	Install *InstallState `json:"install"`

	// loaded from the database but not part of the state itself
	StaleAt *time.Time `json:"stale_at"`
}

func New() *State {
	return &State{}
}

func (i State) AsMap() (map[string]interface{}, error) {
	byts, err := json.Marshal(i)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to json")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(byts, &obj); err != nil {
		return nil, errors.Wrap(err, "unable to convert to map[string]interface{}")
	}

	return obj, nil
}
