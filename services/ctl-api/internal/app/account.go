package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
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
	ID        string                `gorm:"primarykey" json:"id"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_email_subject,unique"`

	Email       string      `json:"email" gorm:"index:idx_email_subject,unique,not null;default null"`
	Subject     string      `json:"subject" gorm:"index:idx_email_subject,unique,not null;"`
	AccountType AccountType `json:"account_type"`

	Roles  []Role  `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"roles"`
	Tokens []Token `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	// ReadOnly Fields
	OrgIDs         []string        `json:"org_ids" gorm:"-"`
	Orgs           []Org           `json:"-" gorm:"-"`
	AllPermissions permissions.Set `json:"permissions" gorm:"-"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountID()
	}

	return nil
}

func (a *Account) AfterQuery(tx *gorm.DB) error {
	a.Orgs = make([]Org, 0)
	a.OrgIDs = make([]string, 0)
	a.AllPermissions = permissions.NewSet()

	visited := make(map[string]struct{}, 0)
	for _, role := range a.Roles {
		for _, policy := range role.Policies {
			a.AllPermissions.Add(policy.Permissions)
		}

		if role.Org == nil {
			continue
		}

		// TODO(jm): this is all pretty messy, a much better approach would be to get the unique org ids from
		// the permission set. This works for now, though.
		if _, ok := visited[role.Org.ID]; ok {
			continue
		}

		a.Orgs = append(a.Orgs, *role.Org)
		a.OrgIDs = append(a.OrgIDs, role.Org.ID)
		visited[role.Org.ID] = struct{}{}
	}

	return nil
}
