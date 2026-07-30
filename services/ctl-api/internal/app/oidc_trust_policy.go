package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

const (
	OIDCTrustPolicyDefaultTokenDuration = 3600
	OIDCTrustPolicyMaxTokenDuration     = 86400
	OIDCTrustPolicyMaxPatternLength     = 512
)

// OIDCTrustPolicy defines a workload identity federation rule for an org: an
// OIDC token whose issuer, audience, and claims match the policy can be
// exchanged for a short-lived Nuon API token bound to the policy's service
// account.
type OIDCTrustPolicy struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"not null" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org   `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	Name    string `json:"name,omitzero" gorm:"not null" temporaljson:"name,omitzero,omitempty"`
	Enabled bool   `json:"enabled" gorm:"default:true" temporaljson:"enabled,omitempty"`

	// IssuerURL is the exact `iss` claim value and the base URL used for OIDC
	// discovery + JWKS fetching. It is always the stored, admin-configured
	// value — never taken from the presented token.
	IssuerURL string `json:"issuer_url,omitzero" gorm:"not null" temporaljson:"issuer_url,omitzero,omitempty"`
	Audience  string `json:"audience,omitzero" gorm:"not null" temporaljson:"audience,omitzero,omitempty"`

	// ClaimConditions maps claim names to patterns. All conditions must match
	// for the policy to apply. Patterns are exact strings, or globs where `*`
	// does not cross `:` segments.
	ClaimConditions map[string]string `json:"claim_conditions,omitzero" gorm:"type:jsonb;serializer:json;default:'{}'" temporaljson:"claim_conditions,omitzero,omitempty"`

	Role                 string `json:"role,omitzero" gorm:"not null" temporaljson:"role,omitzero,omitempty"`
	TokenDurationSeconds int    `json:"token_duration_seconds,omitzero" temporaljson:"token_duration_seconds,omitzero,omitempty"`

	ServiceAccountID string   `json:"service_account_id,omitzero" temporaljson:"service_account_id,omitzero,omitempty"`
	ServiceAccount   *Account `faker:"-" json:"-" temporaljson:"service_account,omitzero,omitempty"`

	LastUsedAt *time.Time `json:"last_used_at,omitempty" temporaljson:"last_used_at,omitzero,omitempty"`
}

func (p *OIDCTrustPolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = domains.NewOIDCTrustPolicyID()
	}

	return nil
}

func (p *OIDCTrustPolicy) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &OIDCTrustPolicy{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
