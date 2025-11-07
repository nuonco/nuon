package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type JobComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// value
	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`

	// Image attributes, copied from a docker_buid or external_image component.
	ImageURL string         `json:"image_url,omitzero" gorm:"notnull" temporaljson:"image_url,omitzero,omitempty"`
	Tag      string         `json:"tag,omitzero" gorm:"notnull" temporaljson:"tag,omitzero,omitempty"`
	Cmd      pq.StringArray `json:"cmd,omitzero" gorm:"type:text[]" temporaljson:"cmd,omitzero,omitempty"`
	EnvVars  pgtype.Hstore  `json:"env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Args     pq.StringArray `json:"args,omitzero" gorm:"type:text[]" swaggertype:"array,string" temporaljson:"args,omitzero,omitempty"`
}

func (j *JobComponentConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: "idx_job_component_config_org_id",
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (e *JobComponentConfig) BeforeCreate(tx *gorm.DB) error {
	e.ID = domains.NewComponentID()
	e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	e.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
