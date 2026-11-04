package app

import (
	"fmt"
	"maps"
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
	PolicyNameRunnerJob  PolicyName = "runner_job"

	// policy names for service accounts
	PolicyNameHostedInstaller    PolicyName = "hosted_installer"
	PolicyNameStack              PolicyName = "stack"
	PolicyNameTelemetryCollector PolicyName = "telemetry_collector"
	PolicyNamePhoneHome          PolicyName = "phone_home"
	PolicyNameLogStream          PolicyName = "log_stream"
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
	Permissions         pgtype.Hstore `json:"-" gorm:"type:hstore" temporaljson:"permissions,omitzero,omitempty"`
	ResponsePermissions pgtype.Hstore `json:"permissions" gorm:"-" swaggertype:"object,string" temporaljson:"-"`
	Document            []byte        `json:"-" gorm:"type:jsonb" temporaljson:"document,omitzero,omitempty"`
}

func (a *Policy) DecodeDocument() (permissions.Policy, error) {
	if a.Document == nil {
		return permissions.Policy{}, fmt.Errorf("policy %s has no document", a.ID)
	}
	if len(a.Permissions) != 0 {
		return permissions.Policy{}, fmt.Errorf("policy %s has both document and legacy permissions", a.ID)
	}
	policy, err := permissions.DecodePolicy(permissions.PolicyFormatV3, a.Document)
	if err != nil {
		return permissions.Policy{}, fmt.Errorf("policy %s: %w", a.ID, err)
	}
	return policy, nil
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
	var err error
	a.ResponsePermissions, err = a.PermissionProjection()
	return err
}

func (a *Policy) BeforeSave(tx *gorm.DB) error {
	var err error
	a.ResponsePermissions, err = a.PermissionProjection()
	return err
}

func (a *Policy) PermissionProjection() (pgtype.Hstore, error) {
	if a.Document == nil {
		return maps.Clone(a.Permissions), nil
	}
	document, err := a.DecodeDocument()
	if err != nil {
		return nil, err
	}
	orgID := a.OrgID.String
	grants, err := document.Compile(orgID, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID})
	if err != nil {
		return nil, err
	}
	sources, err := permissions.LegacySources(grants)
	if err != nil {
		return nil, err
	}
	projection := make(pgtype.Hstore)
	if document.API() != "" {
		projection[permissions.APIPermissionKey] = generics.ToPtr(string(document.API()))
	}
	for object, source := range sources {
		projection[object] = source.Permission.ToStrPtr()
	}
	return projection, nil
}
