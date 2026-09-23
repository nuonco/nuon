package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

const (
	AppConfigRefByTypeInstallWorkflows          = "install_workflows"
	AppConfigRefByTypeInstallStackVersions      = "install_stack_versions"
	AppConfigRefByTypeInstallSandboxRuns        = "install_sandbox_runs"
	AppConfigRefByTypeInstallDeploys            = "install_deploys"
	AppConfigRefByTypeInstallActionWorkflowRuns = "install_action_workflow_runs"
)

type AppConfigRef struct {
	ExpectedConfigID    string     `json:"expected_config_id,omitempty"`
	AppliedConfigID     string     `json:"applied_config_id,omitempty"`
	AppliedConfigAt     *time.Time `json:"applied_config_at,omitempty"`
	AppliedConfigByType string     `json:"applied_config_by_type,omitempty"`
	AppliedConfigByID   string     `json:"applied_config_by_id,omitempty"`
}

func (r *AppConfigRef) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.Errorf("cannot scan %T into AppConfigRef", value)
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, r)
}

func (r AppConfigRef) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (AppConfigRef) GormDataType() string {
	return "jsonb"
}
