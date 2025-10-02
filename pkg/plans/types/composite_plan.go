package plantypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type CompositePlan struct {
	BuildPlan             *BuildPlan             `json:"build_plan,omitempty"`
	DeployPlan            *DeployPlan            `json:"deploy_plan,omitempty"`
	ActionWorkflowRunPlan *ActionWorkflowRunPlan `json:"action_workflow_run_plan,omitempty"`
	SyncSecretsPlan       *SyncSecretsPlan       `json:"sync_secrets_plan,omitempty"`
	SyncOCIPlan           *SyncOCIPlan           `json:"sync_oci_plan,omitempty"`
	SandboxRunPlan        *SandboxRunPlan        `json:"sandbox_run_plan,omitempty"`
}

func (cp CompositePlan) Value() (driver.Value, error) {
	if cp.IsEmpty() {
		return nil, nil
	}
	return json.Marshal(cp)
}

func (cp CompositePlan) IsEmpty() bool {
	return cp.BuildPlan == nil &&
		cp.DeployPlan == nil &&
		cp.ActionWorkflowRunPlan == nil &&
		cp.SyncSecretsPlan == nil &&
		cp.SyncOCIPlan == nil &&
		cp.SandboxRunPlan == nil
}

func (cp *CompositePlan) Scan(value any) error {
	if value == nil {
		*cp = CompositePlan{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into CompositePlan", value)
	}

	if len(bytes) == 0 {
		*cp = CompositePlan{}
		return nil
	}

	return json.Unmarshal(bytes, cp)
}

// GormDataType tells GORM what database type to use
func (CompositePlan) GormDataType() string {
	return "jsonb"
}

// GormDBDataType returns the database data type based on the current using database
func (CompositePlan) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "JSONB"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}
