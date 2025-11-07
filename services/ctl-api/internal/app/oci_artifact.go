package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type OCIArtifact struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;uniqueIndex:idx_owner" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;uniqueIndex:idx_owner" temporaljson:"owner_type,omitzero,omitempty"`

	Tag          string         `json:"tag,omitzero" temporaljson:"tag,omitzero,omitempty"`
	Repository   string         `json:"repository,omitzero" temporaljson:"repository,omitzero,omitempty"`
	MediaType    string         `json:"media_type,omitzero" temporaljson:"media_type,omitzero,omitempty"`
	Digest       string         `json:"digest,omitzero" temporaljson:"digest,omitzero,omitempty"`
	Size         int64          `json:"size,omitzero" gorm:"type:bigint" temporaljson:"size,omitzero,omitempty"`
	URLs         pq.StringArray `gorm:"type:text[]" json:"urls,omitzero" swaggertype:"array,string" temporaljson:"urls,omitzero,omitempty"`
	Annotations  pgtype.Hstore  `json:"annotations,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
	ArtifactType string         `json:"artifact_type,omitzero" gorm:"type:text" temporaljson:"artifact_type,omitzero,omitempty"`

	// Platform fields
	OS           string         `json:"os,omitzero" gorm:"type:text" temporaljson:"os,omitzero,omitempty"`
	Architecture string         `json:"architecture,omitzero" gorm:"type:text" temporaljson:"architecture,omitzero,omitempty"`
	Variant      string         `json:"variant,omitzero" gorm:"type:text" temporaljson:"variant,omitzero,omitempty"`
	OSVersion    string         `json:"os_version,omitzero" gorm:"type:text" temporaljson:"os_version,omitzero,omitempty"`
	OSFeatures   pq.StringArray `gorm:"type:text[]" json:"os_features,omitzero" swaggertype:"array,string" temporaljson:"os_features,omitzero,omitempty"`
}

func (r *OCIArtifact) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &OCIArtifact{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *OCIArtifact) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = domains.NewTerraformWorkspaceID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
