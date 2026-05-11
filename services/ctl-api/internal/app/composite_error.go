package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// CompositeError is the persisted row for a typed, classified error attached
// polymorphically to an owner (workflow_step, runner_job_execution_result,
// component_build, install_deploy, …).
//
// See specs/composite-errors.md for design rationale.
//
// The Type column is the catalog identifier; the Data column is the JSON
// representation of the typed Go struct that implements composite_error.CompositeError.
// The composite_errors helper (Hydrate) uses the catalog to round-trip Data
// back into a typed instance.
type CompositeError struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string                `gorm:"not null;default:null" json:"created_by_id,omitzero"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `gorm:"notnull" json:"created_at,omitzero"`
	UpdatedAt   time.Time             `gorm:"notnull" json:"updated_at,omitzero"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// OrgID is required for RLS. Populated from context in BeforeCreate when
	// not already set by the caller.
	OrgID string `gorm:"not null" json:"org_id,omitzero"`
	Org   Org    `json:"-"`

	// Polymorphic owner — same pattern as QueueSignal. The owner declares a
	// matching `gorm:"polymorphic:Owner;polymorphicValue:<table>"` association
	// to expose CompositeErrors []CompositeError on its model.
	OwnerID   string `gorm:"type:text;check:owner_id_checker,char_length(owner_id)=26;not null" json:"owner_id,omitzero"`
	OwnerType string `gorm:"type:text;not null" json:"owner_type,omitzero"`

	// Catalog classification.
	Type     composite_error.Type     `gorm:"type:text;not null" json:"type,omitzero"`
	Domain   composite_error.Domain   `gorm:"type:text;not null" json:"domain,omitzero"`
	Severity composite_error.Severity `gorm:"type:text;not null" json:"severity,omitzero"`

	// SchemaVersion lets us evolve a type's Data shape without breaking old
	// rows. Defaults to 1; bumped per-type when the Go struct changes
	// incompatibly.
	SchemaVersion int `gorm:"not null;default:1" json:"schema_version,omitzero"`

	// Data is the JSON serialization of the typed CompositeError struct.
	// Round-trips via composite_error/catalog.Hydrate(Type, Data).
	Data json.RawMessage `gorm:"type:jsonb;not null" json:"data,omitempty" swaggertype:"object"`

	// Source is the small parser-input snippet (capped at SourceSnippetMax)
	// plus parser identification, kept so we can debug parser decisions later.
	Source composite_error.Source `gorm:"type:jsonb" json:"source,omitempty"`

	// References point at other entities (or URLs) the UI can dereference at
	// read time — log streams, plan results, runbook links, etc.
	References composite_error.References `gorm:"type:jsonb" json:"references,omitempty"`

	// Title/SummaryCached are denormalized copies of Render() output at write
	// time. Used only for cheap admin search / listing — canonical render is
	// always re-derived from the typed instance.
	TitleCached   string `gorm:"type:text" json:"title_cached,omitzero"`
	SummaryCached string `gorm:"type:text" json:"summary_cached,omitzero"`

	// Render is the user-facing payload computed by the typed CompositeError's
	// Render() method. Populated lazily via AfterFind from Type+Data so API
	// responses can render rich error views without the dashboard having to
	// hydrate the typed value itself. Not persisted.
	Render *composite_error.Render `gorm:"-" json:"render,omitempty"`
}

func (r *CompositeError) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &CompositeError{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &CompositeError{}, "owner_type_owner_id_deleted_at"),
			Columns: []string{"owner_type", "owner_id", "deleted_at"},
		},
		{
			Name:    indexes.Name(db, &CompositeError{}, "type"),
			Columns: []string{"type"},
		},
		{
			Name:    indexes.Name(db, &CompositeError{}, "domain"),
			Columns: []string{"domain"},
		},
		{
			Name:    indexes.Name(db, &CompositeError{}, "severity"),
			Columns: []string{"severity"},
		},
		{
			Name:    indexes.Name(db, &CompositeError{}, "created_at"),
			Columns: []string{"created_at"},
		},
	}
}

func (r *CompositeError) TableName() string { return "composite_errors" }

func (r *CompositeError) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewCompositeErrorID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if r.SchemaVersion == 0 {
		r.SchemaVersion = 1
	}
	return nil
}

// AfterFind hydrates the typed instance via the in-memory catalog and stores
// the Render() output on the row so it ships in the JSON response. Errors
// (unknown type, malformed Data) are swallowed: the row still serializes
// with title_cached / summary_cached as a fallback.
func (r *CompositeError) AfterFind(tx *gorm.DB) error {
	val, err := catalog.Hydrate(r.Type, r.Data)
	if err != nil || val == nil {
		return nil
	}
	rendered := val.Render(tx.Statement.Context)
	r.Render = &rendered
	return nil
}
