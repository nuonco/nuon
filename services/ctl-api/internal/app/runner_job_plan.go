package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerJobPermissionTraceRecord struct {
	RoleName   string `json:"role_name,omitempty"`
	RoleSource string `json:"role_source,omitempty"`
	Available  bool   `json:"available,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
	Selected   bool   `json:"selected,omitempty"`
}

type RunnerJobPermissionInfo struct {
	Role               string                           `json:"role,omitempty"`
	RoleSource         string                           `json:"role_source,omitempty"`
	RoleSelectionTrace []RunnerJobPermissionTraceRecord `json:"role_selection_trace,omitempty"`
}

func (r *RunnerJobPermissionInfo) Scan(value interface{}) error {
	if value == nil {
		*r = RunnerJobPermissionInfo{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into RunnnerJobPermissionInfo", value)
	}

	if len(bytes) == 0 {
		*r = RunnerJobPermissionInfo{}
		return nil
	}

	return json.Unmarshal(bytes, r)
}

func (r RunnerJobPermissionInfo) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type RunnerJobPlan struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_job_plan,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobID string `json:"runner_job_id,omitzero" gorm:"defaultnull;notnull;index:idx_runner_job_plan,unique" temporaljson:"runner_job_id,omitzero,omitempty"`

	PermissionInfo RunnerJobPermissionInfo `json:"permission_info,omitzero" gorm:"type:jsonb" temporaljson:"permission_info"`

	PlanJSON      string                  `json:"plan_json,omitzero" temporaljson:"plan_json,omitzero,omitempty"`
	CompositePlan plantypes.CompositePlan `json:"composite_plan,omitzero" gorm:"type:jsonb" temporaljson:"composite_plan,omitzero,omitempty"`
}

func (r *RunnerJobPlan) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJobPlan{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *RunnerJobPlan) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
