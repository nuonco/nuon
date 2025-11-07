package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type LogStream struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Open bool `json:"open,omitzero" temporaljson:"open,omitzero,omitempty"`

	Attrs pgtype.Hstore `json:"attrs,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"attrs,omitzero,omitempty"`

	ParentLogStreamID generics.NullString `json:"-" swaggerignore:"true" temporaljson:"parent_log_stream_id,omitzero,omitempty"`
	ParentLogStream   *LogStream          `json:"-" faker:"-" temporaljson:"parent_log_stream,omitzero,omitempty"`

	RunnerJobs []RunnerJob `json:"-" temporaljson:"runner_jobs,omitzero,omitempty"`

	// fields not stored in the DB

	WriteToken   string `json:"write_token,omitzero" gorm:"-" temporaljson:"write_token,omitzero,omitempty"`
	RunnerAPIURL string `json:"runner_api_url,omitzero" gorm:"-" temporaljson:"runner_apiurl,omitzero,omitempty"`
}

func (r *LogStream) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewLogStreamID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *LogStream) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &LogStream{}, "preload"),
			Columns: []string{
				"owner_id",
				"owner_type",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &LogStream{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
