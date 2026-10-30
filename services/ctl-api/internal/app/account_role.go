package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AccountRole is a many2many table used by gorm under the hood
type AccountRole struct {
	ID string `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_account_role:unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID generics.NullString `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org                `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	RoleID string `gorm:"primary_key;index:idx_account_role:unique" temporaljson:"role_id,omitzero,omitempty"`
	Role   Role   `temporaljson:"role,omitzero,omitempty"`

	AccountID string  `json:"account_id,omitzero" gorm:"primary_key;index:idx_account_role:unique" temporaljson:"account_id,omitzero,omitempty"`
	Account   Account `json:"account,omitzero" temporaljson:"account,omitzero,omitempty"`

	ScopeKind permissions.ResourceKind `json:"-" temporaljson:"scope_kind,omitzero,omitempty"`
	ScopeID   string                   `json:"-" temporaljson:"scope_id,omitzero,omitempty"`
}

func (c *AccountRole) BeforeSave(tx *gorm.DB) error {
	if c.ScopeKind != "" || c.ScopeID != "" {
		scope := permissions.Scope{Kind: c.ScopeKind, OrgID: c.OrgID.ValueString(), ID: c.ScopeID}
		if err := scope.ValidateShape(); err != nil {
			return err
		}
		if c.ScopeKind != permissions.KindOrg && c.ScopeKind != permissions.KindApp && c.ScopeKind != permissions.KindInstall {
			return fmt.Errorf("role binding requires an org, app or install scope")
		}
	}
	if c.ID == "" {
		c.ID = domains.NewAccountID()
	}

	if c.OrgID.Empty() {
		c.OrgID = generics.NewNullString(orgIDFromContext(tx.Statement.Context))
	}

	return nil
}

func (a *AccountRole) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AccountRole{}, "account_id"),
			Columns: []string{
				"account_id",
			},
		},
		{
			Name: indexes.Name(db, &AccountRole{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AccountRole{}, "org_id_deleted_at_account_id"),
			Columns: []string{
				"org_id",
				"deleted_at",
				"account_id",
			},
		},
	}
}
