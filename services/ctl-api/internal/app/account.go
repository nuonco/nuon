package app

import (
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AccountType string

const (
	AccountTypeAuth    AccountType = "auth"
	AccountTypeAuth0   AccountType = "auth0"
	AccountTypeService AccountType = "service"

	// Internal Account Types for testing
	AccountTypeCanary      AccountType = "canary"
	AccountTypeIntegration AccountType = "integration"
)

type Account struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_email_subject,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	Email       string      `json:"email,omitzero" gorm:"index:idx_email_subject,unique,not null;default null" temporaljson:"email,omitzero,omitempty"`
	Subject     string      `json:"subject,omitzero" gorm:"index:idx_email_subject,unique,not null;" temporaljson:"subject,omitzero,omitempty"`
	Name        string      `json:"name,omitzero" gorm:"default:null" temporaljson:"name,omitzero,omitempty"`
	AccountType AccountType `json:"account_type,omitzero" temporaljson:"account_type,omitzero,omitempty"`

	Roles        []Role            `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"roles,omitzero" temporaljson:"roles,omitzero,omitempty"`
	Grants       []ResourceGrant   `gorm:"constraint:OnDelete:CASCADE;" json:"grants,omitzero" temporaljson:"grants,omitzero,omitempty"`
	Tokens       []Token           `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"tokens,omitzero,omitempty"`
	Identities   []AccountIdentity `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"identities,omitzero,omitempty"`
	UserJourneys UserJourneys      `json:"user_journeys,omitzero" gorm:"type:jsonb;default null" temporaljson:"user_journeys,omitzero,omitempty"`

	// ReadOnly Fields
	OrgIDs         []string        `json:"org_ids,omitzero" gorm:"-" temporaljson:"org_i_ds,omitzero,omitempty"`
	Orgs           []*Org          `json:"-" gorm:"-" temporaljson:"orgs,omitzero,omitempty"`
	AllPermissions permissions.Set `json:"permissions,omitzero" gorm:"-" temporaljson:"all_permissions,omitzero,omitempty"`

	IsEmployee bool `json:"-"`
}

func (a *Account) Indexes(db *gorm.DB) []migrations.Index {

	return []migrations.Index{
		{
			Name: indexes.Name(db, &Account{}, "email"),
			Columns: []string{
				"email",
				"deleted_at",
			},
		},
		{Name: indexes.Name(db, &Account{}, "subject"),
			Columns: []string{
				"subject",
				"deleted_at",
			},
		},
	}
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountID()
	}

	return nil
}

func (a *Account) AfterQuery(tx *gorm.DB) error {
	a.IsEmployee = slices.Contains([]AccountType{AccountTypeAuth0, AccountTypeAuth}, a.AccountType) && strings.HasSuffix(a.Email, "@nuon.co")

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

	for i := range a.Grants {
		grant := a.Grants[i]

		if perm, err := permissions.NewPermission(grant.Permission); err == nil {
			a.AllPermissions.Grant(grant.ResourceID, perm)
		}

		if grant.OrgID == "" {
			continue
		}
		if _, ok := visited[grant.OrgID]; ok {
			continue
		}

		a.OrgIDs = append(a.OrgIDs, grant.OrgID)
		if grant.Org.ID != "" {
			a.Orgs = append(a.Orgs, &a.Grants[i].Org)
		}
		visited[grant.OrgID] = struct{}{}
	}

	return nil
}

// HasOrg reports whether the account has any access to the given org — via an
// org role or a resource grant. This is membership, not authorization: it does
// not imply any particular permission, only that org-scoped enforcement should
// proceed to a downstream (grant-level) check rather than reject outright.
func (a *Account) HasOrg(orgID string) bool {
	return slices.Contains(a.OrgIDs, orgID)
}

func (*Account) JoinTables() []migrations.JoinTable {
	return []migrations.JoinTable{
		{
			Field:     "Roles",
			JoinTable: &AccountRole{},
		},
	}
}
