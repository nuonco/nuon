package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type LogStream struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id"`
	Org   Org    `json:"-"`

	OwnerID   string `json:"owner_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26"`
	OwnerType string `json:"owner_type" gorm:"type:text;"`

	Open bool `json:"open"`

	Attrs pgtype.Hstore `json:"attrs" gorm:"type:hstore" swaggertype:"object,string"`

	ParentLogStreamID generics.NullString `json:"-" swaggerignore:"true"`
	ParentLogStream   *LogStream          `json:"-" faker:"-"`

	RunnerJobs []RunnerJob `json:"-"`

	// fields not stored in the DB

	WriteToken   string `json:"write_token" temporaljson:"write_token" gorm:"-"`
	RunnerAPIURL string `json:"runner_api_url" temporaljson:"runner_api_url" gorm:"-"`
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
