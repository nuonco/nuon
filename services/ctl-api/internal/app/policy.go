package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type PolicyName string

const (
	// we create a custom policy for each role
	PolicyNameOrgAdmin  PolicyName = "org_admin"
	PolicyNameInstaller PolicyName = "installer"
	PolicyNameRunner    PolicyName = "runner"
)

type Policy struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"notnull;index:idx_org_role_policy,unique"`

	RoleID string `json:"role_id" gorm:"notnull;default null"`
	Role   Role   `swaggerignore:"true" json:"role"`

	OrgID string `json:"org_id" gorm:"notnull;index:idx_org_role_policy,unique" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Name PolicyName `json:"name" gorm:"index:idx_org_role_policy,unique"`

	// Permissions are used to track granular permissions for each domain
	Permissions pgtype.Hstore `json:"permissions" gorm:"type:hstore" swaggertype:"object,string"`
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
