package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type JobComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	// value
	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-"`

	// Image attributes, copied from a docker_buid or external_image component.
	ImageURL string         `json:"image_url" gorm:"notnull"`
	Tag      string         `json:"tag" gorm:"notnull"`
	Cmd      pq.StringArray `json:"cmd" gorm:"type:text[]"`
	EnvVars  pgtype.Hstore  `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Args     pq.StringArray `json:"args" gorm:"type:text[]" swaggertype:"array,string"`
}

func (e *JobComponentConfig) BeforeCreate(tx *gorm.DB) error {
	e.ID = domains.NewComponentID()
	e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	e.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
