package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// GrantResourceType identifies the tier of resource a ResourceGrant scopes to.
// The initial scope is the org -> app -> install spine plus delegable org-owned
// siblings (webhooks, VCS connections, Slack subscriptions); deeper leaf targets
// (component, install-component) and further siblings (runner, etc.) are added later.
type GrantResourceType string

const (
	GrantResourceTypeOrg     GrantResourceType = "org"
	GrantResourceTypeApp     GrantResourceType = "app"
	GrantResourceTypeInstall GrantResourceType = "install"

	// org-owned siblings: chains are [entity, org], with no intermediate tier.
	GrantResourceTypeWebhook           GrantResourceType = "webhook"
	GrantResourceTypeVCSConnection     GrantResourceType = "vcs_connection"
	GrantResourceTypeSlackSubscription GrantResourceType = "slack_subscription"
)

// GrantResourceWildcard is a ResourceID that scopes a grant to every resource
// of its ResourceType within the grant's org (e.g. install + "*" == all installs
// in the org). It is deliberately distinct from the type-blind permissions.Set
// "*" wildcard, which would match every object regardless of tier.
const GrantResourceWildcard = "*"

// IsWildcard reports whether the grant covers all resources of its type in the org.
func (r ResourceGrant) IsWildcard() bool {
	return r.ResourceID == GrantResourceWildcard
}

// ResourceGrant grants an account a permission on a single resource. It is
// folded into the account's AllPermissions alongside role policies, and
// authorization walks up the resource's ownership chain (resource -> ... -> org)
// looking for a grant that satisfies the requested permission.
type ResourceGrant struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_resource_grant_unique,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// org the grant lives in; drives membership resolution (account OrgIDs)
	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_resource_grant_unique,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"org,omitzero,omitempty"`

	// account the grant is issued to
	AccountID string  `json:"account_id,omitzero" gorm:"notnull;index:idx_resource_grant_unique,unique" temporaljson:"account_id,omitzero,omitempty"`
	Account   Account `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"account,omitzero,omitempty"`

	ResourceType GrantResourceType `json:"resource_type,omitzero" gorm:"notnull;default null;index:idx_resource_grant_unique,unique" temporaljson:"resource_type,omitzero,omitempty"`
	ResourceID   string            `json:"resource_id,omitzero" gorm:"notnull;default null;index:idx_resource_grant_unique,unique" temporaljson:"resource_id,omitzero,omitempty"`
	Permission   string            `json:"permission,omitzero" gorm:"notnull;default null" temporaljson:"permission,omitzero,omitempty"`
}

func (r *ResourceGrant) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &ResourceGrant{}, "account_id"),
			Columns: []string{"account_id"},
		},
		{
			Name:    indexes.Name(db, &ResourceGrant{}, "resource_id"),
			Columns: []string{"resource_id"},
		},
	}
}

func (r *ResourceGrant) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewResourceGrantID()
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
