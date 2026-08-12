package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type PolicyName string

const (
	// we create a custom policy for each role
	PolicyNameOrgAdmin    PolicyName = "org_admin"
	PolicyNameOrgSupport  PolicyName = "org_support"
	PolicyNameOrgReadOnly PolicyName = "org_read_only"
	// Deprecated: the org_builder role is no longer created or assignable.
	PolicyNameOrgBuilder PolicyName = "org_builder"
	PolicyNameInstaller  PolicyName = "installer"
	PolicyNameRunner     PolicyName = "runner"

	// policy names for service accounts
	PolicyNameHostedInstaller PolicyName = "hosted_installer"

	// PolicyNameCustom is the single policy carried by an org-defined custom role.
	PolicyNameCustom PolicyName = "custom"
)

// Level enumerates the resource kinds authz recognizes — the link types an
// ownership chain can contain, and the kinds a scoped permission can target
// or be confined to.
type Level string

const (
	LevelOrg       Level = "org"
	LevelApp       Level = "app"
	LevelInstall   Level = "install"
	LevelAppBranch Level = "app_branch"
)

func NewLevel(val string) (Level, error) {
	switch Level(val) {
	case LevelOrg, LevelApp, LevelInstall, LevelAppBranch:
		return Level(val), nil
	}
	return "", fmt.Errorf("invalid resource type %q", val)
}

// ValidWildcardScope reports whether a "*" entry on resources of type l may
// be confined to a parent of type scope. The empty scope (org-wide) is always
// legal and is the only way to express an org-wide wildcard — org is never
// named explicitly.
func (l Level) ValidWildcardScope(scope Level) bool {
	if scope == "" {
		return true
	}
	switch l {
	case LevelInstall, LevelAppBranch:
		return scope == LevelApp
	}
	return false
}

// PermissionEntry is one scoped permission on a policy: verbs on a single
// resource, or on all resources of a type (ResourceID "*") optionally
// confined to a parent scope.
type PermissionEntry struct {
	ResourceType Level             `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	ScopeType    Level             `json:"scope_type,omitzero"`
	ScopeID      string            `json:"scope_id,omitzero"`
	Permissions  permissions.Verbs `json:"permissions" swaggertype:"array,string"`
}

// TypeGrant is a resolved wildcard entry: verbs on every resource of a type,
// confined to ScopeID when set (org-wide when empty).
type TypeGrant struct {
	ScopeID string
	Verbs   permissions.Verbs
}

// TypeGrants indexes an account's wildcard grants by org id, then resource
// type.
type TypeGrants map[string]map[Level][]TypeGrant

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

	// ScopedPermissions carry permissions confined to a single resource or a
	// type-wildcard under a parent scope; org-tier permissions stay in the
	// Permissions hstore untouched.
	ScopedPermissions []PermissionEntry `json:"scoped_permissions,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"scoped_permissions,omitzero,omitempty"`
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
