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

type PolicyName string

const (
	// we create a custom policy for each role
	PolicyNameOrgAdmin  PolicyName = "org_admin"
	PolicyNameInstaller PolicyName = "installer"
	PolicyNameRunner    PolicyName = "runner"

	// policy names for service accounts
	PolicyNameHostedInstaller PolicyName = "hosted_installer"
)

type Policy struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	RoleID string `json:"role_id,omitzero" gorm:"notnull;default null" temporaljson:"role_id,omitzero,omitempty"`
	Role   Role   `swaggerignore:"true" json:"role,omitzero" temporaljson:"role,omitzero,omitempty"`

	OrgID generics.NullString `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org                `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Name PolicyName `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`

	// Permissions are used to track granular permissions for each domain
	Permissions pgtype.Hstore `json:"permissions" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"permissions,omitzero,omitempty"`
}

func (a *Policy) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Policy{}, "role_id"),
			Columns: []string{
				"role_id",
			},
			UniqueValue: generics.NewNullBool(true),
		},
	}
}

func (a *Policy) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountPolicyID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *Policy) AfterQuery(tx *gorm.DB) error {
	return nil
}
