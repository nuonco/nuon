package app

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type NotebookStatus string

const (
	NotebookStatusActive   NotebookStatus = "active"
	NotebookStatusArchived NotebookStatus = "archived"
)

// Notebook is an install-scoped, Jupyter-style execution surface. Each cell
// runs a command on the install's runner via a long-lived, warm per-notebook
// Temporal workflow. The notebook itself is the durable resource; cell content
// lives in NotebookCell and execution history in NotebookCellRun.
type Notebook struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	Name        string         `json:"name,omitzero" gorm:"notnull;default:''" temporaljson:"name,omitzero,omitempty"`
	Description string         `json:"description,omitzero" temporaljson:"description,omitzero,omitempty"`
	Status      NotebookStatus `json:"status,omitzero" gorm:"notnull;default:'active'" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`

	Cells []NotebookCell `json:"cells,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"cells,omitzero,omitempty"`

	// Queue owns the lifecycle of the notebook's warm Temporal workflow: a
	// notebook-start signal enqueued at create time starts NotebookWorkflow,
	// and the queue can re-dispatch it for recovery. Cell runs still dispatch
	// to the workflow directly via update-with-start.
	Queue Queue `json:"queue,omitzero" gorm:"polymorphic:Owner;" temporaljson:"queue,omitzero,omitempty"`
}

func (n *Notebook) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &Notebook{}, "install"),
			Columns: []string{"install_id", "deleted_at", "updated_at DESC"},
		},
		{
			Name:    indexes.Name(db, &Notebook{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}

func (n *Notebook) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = domains.NewNotebookID()
	}
	if n.CreatedByID == "" {
		n.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if n.OrgID == "" {
		n.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if n.Status == "" {
		n.Status = NotebookStatusActive
	}
	return nil
}

// NotebookCell is an editable, runnable cell within a Notebook. Add/edit/delete
// /reorder of cells are plain DB writes and never touch the workflow; only
// running a cell does. Editing a cell bumps Revision so the UI can flag a cell
// as "edited since last run".
type NotebookCell struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	NotebookID string   `json:"notebook_id,omitzero" gorm:"not null;default null" temporaljson:"notebook_id,omitzero,omitempty"`
	Notebook   Notebook `swaggerignore:"true" json:"-" temporaljson:"notebook,omitzero,omitempty"`

	// Position is the 0-based ordering of this cell within the notebook.
	Position int `json:"position" gorm:"notnull;default:0" temporaljson:"position,omitzero,omitempty"`

	// Revision increments on every edit; runs snapshot the revision they ran.
	Revision int `json:"revision" gorm:"notnull;default:1" temporaljson:"revision,omitzero,omitempty"`

	Name             string        `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`
	InlineContents   string        `json:"inline_contents,omitzero" temporaljson:"inline_contents,omitzero,omitempty"`
	Command          string        `json:"command,omitzero" temporaljson:"command,omitzero,omitempty"`
	EnvVars          pgtype.Hstore `json:"env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Timeout          time.Duration `json:"timeout,omitzero" gorm:"default:0;not null" swaggertype:"primitive,integer" temporaljson:"timeout,omitzero,omitempty"`
	Role             string        `json:"role,omitzero" temporaljson:"role,omitzero,omitempty"`
	EnableKubeConfig sql.NullBool  `json:"enable_kube_config" gorm:"default:true" temporaljson:"enable_kube_config"`

	// LatestRun is populated on read so the UI can show the most recent run's
	// status and log stream directly below the cell. Not persisted.
	LatestRun *NotebookCellRun `json:"latest_run,omitzero" gorm:"-" temporaljson:"latest_run,omitzero,omitempty"`
}

func (c *NotebookCell) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &NotebookCell{}, "notebook"),
			Columns: []string{"notebook_id", "deleted_at", "position"},
		},
		{
			Name:    indexes.Name(db, &NotebookCell{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}

func (c *NotebookCell) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = domains.NewNotebookCellID()
	}
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if c.Revision == 0 {
		c.Revision = 1
	}
	return nil
}

// NotebookCellRun is the product/history record for a single execution of a
// cell. It links to the existing execution/audit artifacts rather than
// replacing them: the InstallActionWorkflowRun (and its LogStream/RunnerJob)
// remains the source of truth for execution and audit. The cell config is
// snapshotted so history stays truthful even after the cell is edited or
// deleted.
type NotebookCellRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string `json:"install_id,omitzero" gorm:"not null;default null" temporaljson:"install_id,omitzero,omitempty"`

	NotebookID string `json:"notebook_id,omitzero" gorm:"not null;default null" temporaljson:"notebook_id,omitzero,omitempty"`
	CellID     string `json:"cell_id,omitzero" gorm:"not null;default null" temporaljson:"cell_id,omitzero,omitempty"`
	// CellRevision records which revision of the cell this run executed.
	CellRevision int `json:"cell_revision,omitzero" temporaljson:"cell_revision,omitzero,omitempty"`

	// IdempotencyKey deduplicates run requests (HTTP retries / update retries).
	// A server-side key is always generated, so the composite unique index on
	// (notebook_id, idempotency_key) in Indexes() never sees an empty key.
	IdempotencyKey string `json:"idempotency_key,omitzero" temporaljson:"idempotency_key,omitzero,omitempty"`

	// Link to the existing execution/audit artifacts.
	InstallActionWorkflowRunID string `json:"install_action_workflow_run_id,omitzero" temporaljson:"install_action_workflow_run_id,omitzero,omitempty"`
	LogStreamID                string `json:"log_stream_id,omitzero" temporaljson:"log_stream_id,omitzero,omitempty"`
	RunnerJobID                string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`

	// Cell config snapshot at run time.
	Name           string        `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`
	InlineContents string        `json:"inline_contents,omitzero" temporaljson:"inline_contents,omitzero,omitempty"`
	Command        string        `json:"command,omitzero" temporaljson:"command,omitzero,omitempty"`
	EnvVars        pgtype.Hstore `json:"env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`

	TriggeredByID   string `json:"triggered_by_id,omitzero" temporaljson:"triggered_by_id,omitzero,omitempty"`
	TriggeredByType string `json:"triggered_by_type,omitzero" temporaljson:"triggered_by_type,omitzero,omitempty"`

	Status            InstallActionWorkflowRunStatus `json:"status,omitzero" gorm:"notnull;default:'queued'" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                         `json:"status_description,omitzero" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus                `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`
}

func (r *NotebookCellRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &NotebookCellRun{}, "notebook"),
			Columns: []string{"notebook_id", "deleted_at", "created_at DESC"},
		},
		{
			Name:    indexes.Name(db, &NotebookCellRun{}, "cell"),
			Columns: []string{"cell_id", "deleted_at", "created_at DESC"},
		},
		{
			Name:        indexes.Name(db, &NotebookCellRun{}, "idempotency"),
			Columns:     []string{"notebook_id", "idempotency_key"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name:    indexes.Name(db, &NotebookCellRun{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}

func (r *NotebookCellRun) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewNotebookCellRunID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if r.Status == "" {
		r.Status = InstallActionRunStatusQueued
	}
	return nil
}
