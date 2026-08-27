package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RoleType string

const (
	// user roles
	RoleTypeOrgAdmin RoleType = "org_admin"
	// Deprecated: org_support is no longer assignable to new members; use org_admin or org_read_only.
	RoleTypeOrgSupport  RoleType = "org_support"
	RoleTypeOrgReadOnly RoleType = "org_read_only"
	// Deprecated: org_builder is no longer assignable; grant automation an
	// org_read_only role plus resource grants for anything it must write.
	RoleTypeOrgBuilder RoleType = "org_builder"

	// service account roles
	// Deprecated: installer is no longer assignable; it grants full org access, use org_admin instead.
	RoleTypeInstaller       RoleType = "installer"
	RoleTypeRunner          RoleType = "runner"
	RoleTypeHostedInstaller RoleType = "hosted-installer"
	// RoleTypeStack is per-install: one role per install-stack service account,
	// granting only that install's stack endpoints.
	RoleTypeStack RoleType = "stack"
)

// Role contexts name the assignment surfaces a role may be offered on. A role
// with no contexts still exists (and is displayed where held) but no picker or
// create endpoint offers it.
const (
	RoleContextTeam           = "team"
	RoleContextServiceAccount = "service_account"
	RoleContextAPIToken       = "api_token"
	RoleContextTrustPolicy    = "oidc_trust_policy"
)

type Role struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull;defaultnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `temporaljson:"created_by,omitzero,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Accounts []Account `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"-" temporaljson:"accounts,omitzero,omitempty"`

	// NOTE: not all roles have to belong to an org, this is mainly for historical reasons.
	OrgID generics.NullString `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org                `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	RoleType RoleType `json:"role_type,omitzero" gorm:"defaultnull;notnull" temporaljson:"role_type,omitzero,omitempty"`

	// display + assignability metadata; the single source of truth read by
	// GET /v1/roles and every role picker. Managed roles are kept in sync
	// with standardOrgRoles by the authz reconciler.
	Title       string   `json:"title,omitzero" temporaljson:"title,omitzero,omitempty"`
	Description string   `json:"description,omitzero" temporaljson:"description,omitzero,omitempty"`
	Contexts    []string `json:"applies_to,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"contexts,omitzero,omitempty"`
	Managed     bool     `json:"managed" temporaljson:"managed,omitzero,omitempty"`

	Policies []Policy `json:"policies,omitzero" temporaljson:"policies,omitzero,omitempty"`
}

// AllowsContext reports whether the role may be offered on the given
// assignment surface.
func (a *Role) AllowsContext(roleContext string) bool {
	for _, c := range a.Contexts {
		if c == roleContext {
			return true
		}
	}
	return false
}

func (a *Role) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &Role{}, "org_id"),
			Columns: []string{"org_id", "role_type"},
		},
	}
}

func (a *Role) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewRoleID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *Role) AfterQuery(tx *gorm.DB) error {
	return nil
}
