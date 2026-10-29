package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

const (
	InstallGroupRunPhaseDeploy  = "deploy"
	InstallGroupRunPhaseRunbook = "runbook"
	InstallGroupRunPhasePin     = "pin"
)

type InstallGroupRunInstall struct {
	InstallID  string `json:"install_id"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Status     string `json:"status"`

	// ReleaseReason is why the group stopped waiting on this install.
	ReleaseReason string `json:"release_reason,omitempty"`

	// WaitingOnRunID is the app branch run this install is queued behind.
	WaitingOnRunID string `json:"waiting_on_run_id,omitempty"`

	// SupersededByRunID is the later app branch run that cancelled this one.
	SupersededByRunID string `json:"superseded_by_run_id,omitempty"`

	// Phase is which stage of the group the install is in: "deploy" or "runbook".
	Phase    string                   `json:"phase,omitempty"`
	Runbooks []InstallGroupRunRunbook `json:"runbooks,omitempty"`
}

// InstallGroupRunRunbook records one post-deploy runbook run for an install.
type InstallGroupRunRunbook struct {
	RunbookID   string `json:"runbook_id"`
	RunbookName string `json:"runbook_name"`
	RunID       string `json:"run_id,omitempty"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	Status      string `json:"status"`

	// Attempt increments when a retry of the step re-runs a runbook that failed,
	// so the retry gets a fresh idempotency key instead of adopting the failed run.
	Attempt int `json:"attempt,omitempty"`
}

type InstallGroupRun struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchRunID string       `json:"app_branch_run_id,omitzero" gorm:"not null" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"-" temporaljson:"app_branch_run,omitzero,omitempty"`

	InstallGroupID string                `json:"install_group_id,omitzero" gorm:"not null" temporaljson:"install_group_id,omitzero,omitempty"`
	InstallGroup   AppBranchInstallGroup `faker:"-" json:"install_group,omitempty" temporaljson:"install_group,omitzero,omitempty"`

	InstallGroupName string `json:"install_group_name,omitzero" gorm:"not null" temporaljson:"install_group_name,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Installs []InstallGroupRunInstall `json:"installs,omitempty" gorm:"type:jsonb;serializer:install_group_run_installs" temporaljson:"installs,omitzero,omitempty"`

	TotalInstalls     int `json:"total_installs,omitzero" temporaljson:"total_installs,omitzero,omitempty"`
	CompletedInstalls int `json:"completed_installs,omitzero" temporaljson:"completed_installs,omitzero,omitempty"`
	FailedInstalls    int `json:"failed_installs,omitzero" temporaljson:"failed_installs,omitzero,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty" temporaljson:"completed_at,omitzero,omitempty"`
}

func (i *InstallGroupRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "app_branch_run_id"),
			Columns: []string{
				"app_branch_run_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "install_group_id"),
			Columns: []string{
				"install_group_id",
			},
		},
	}
}

func (i *InstallGroupRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallGroupRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func init() {
	schema.RegisterSerializer("install_group_run_installs", installGroupRunInstallsSerializer{})
}

// installGroupRunInstallsSerializer matches encoding/json for a slice, and also
// accepts a single JSON object. GORM Update of a one-element slice stored that
// object, and the stock json serializer cannot scan it back into the slice.
type installGroupRunInstallsSerializer struct{}

func (installGroupRunInstallsSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	fieldValue := reflect.New(field.FieldType)
	if dbValue != nil {
		var raw []byte
		switch v := dbValue.(type) {
		case []byte:
			raw = v
		case string:
			raw = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal install group run installs: %#v", dbValue)
		}
		if err := unmarshalInstallGroupRunInstalls(raw, fieldValue.Interface()); err != nil {
			return err
		}
	}
	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (installGroupRunInstallsSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	result, err := json.Marshal(fieldValue)
	if string(result) == "null" {
		if field.TagSettings["NOT NULL"] != "" {
			return "", nil
		}
		return nil, err
	}
	return string(result), err
}

func unmarshalInstallGroupRunInstalls(raw []byte, dest any) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if raw[0] == '{' {
		if bytes.Equal(raw, []byte("{}")) {
			raw = []byte("[]")
		} else {
			raw = append(append([]byte{'['}, raw...), ']')
		}
	}
	return json.Unmarshal(raw, dest)
}
