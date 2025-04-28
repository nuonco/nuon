package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AccountType string

const (
	AccountTypeAuth0   AccountType = "auth0"
	AccountTypeService AccountType = "service"

	// Internal Account Types for testing
	AccountTypeCanary      AccountType = "canary"
	AccountTypeIntegration AccountType = "integration"
)

type Account struct {
	ID        string                `gorm:"primarykey" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_email_subject,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	Email       string      `json:"email" gorm:"index:idx_email_subject,unique,not null;default null" temporaljson:"email,omitzero,omitempty"`
	Subject     string      `json:"subject" gorm:"index:idx_email_subject,unique,not null;" temporaljson:"subject,omitzero,omitempty"`
	AccountType AccountType `json:"account_type" temporaljson:"account_type,omitzero,omitempty"`

	Roles  []Role  `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"roles" temporaljson:"roles,omitzero,omitempty"`
	Tokens []Token `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"tokens,omitzero,omitempty"`

	// ReadOnly Fields
	OrgIDs         []string        `json:"org_ids" gorm:"-" temporaljson:"org_i_ds,omitzero,omitempty"`
	Orgs           []*Org          `json:"-" gorm:"-" temporaljson:"orgs,omitzero,omitempty"`
	AllPermissions permissions.Set `json:"permissions" gorm:"-" temporaljson:"all_permissions,omitzero,omitempty"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountID()
	}

	return nil
}

func (a *Account) AfterQuery(tx *gorm.DB) error {
	a.OrgIDs = make([]string, 0)
	a.AllPermissions = permissions.NewSet()

	visited := make(map[string]struct{}, 0)
	for _, role := range a.Roles {
		for _, policy := range role.Policies {
			a.AllPermissions.Add(policy.Permissions)
		}

		if role.OrgID.Empty() {
			continue
		}

		// TODO(jm): this is all pretty messy, a much better approach would be to get the unique org ids from
		// the permission set. This works for now, though.
		if _, ok := visited[role.Org.ID]; ok {
			continue
		}

		a.OrgIDs = append(a.OrgIDs, role.Org.ID)
		a.Orgs = append(a.Orgs, role.Org)
		visited[role.Org.ID] = struct{}{}
	}

	return nil
}

func (*Account) JoinTables() []migrations.JoinTable {
	return []migrations.JoinTable{
		{
			Field:     "Roles",
			JoinTable: &AccountRole{},
		},
	}
}
