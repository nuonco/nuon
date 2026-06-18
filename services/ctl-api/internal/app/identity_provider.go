package app

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ProviderType string

const (
	ProviderTypeOIDC   ProviderType = "oidc"
	ProviderTypeGoogle ProviderType = "google"
	ProviderTypeGitHub ProviderType = "github"
)

type IdentityProvider struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_provider_type,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// OrgID can be nil for global providers available to all orgs in the deployment.
	// If set, the provider is only available to users in that specific org.
	OrgID *string `json:"org_id,omitempty" gorm:"index:idx_provider_type,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org    `faker:"-" json:"-" gorm:"constraint:OnDelete:SET NULL" temporaljson:"org,omitzero,omitempty"`

	ProviderType ProviderType `json:"provider_type,omitzero" gorm:"not null,index:idx_provider_type,unique" temporaljson:"provider_type,omitzero,omitempty"`
	Enabled      bool         `json:"enabled" gorm:"default:false" temporaljson:"enabled,omitempty"`

	// Config holds provider-specific configuration as JSON.
	// The structure depends on ProviderType:
	// - oidc: providers.OpenIDConfig
	// - google: providers.GoogleConfig
	// - github: providers.GitHubConfig
	Config []byte `json:"config,omitzero" gorm:"type:jsonb" temporaljson:"config,omitzero,omitempty"`
}

func (a *IdentityProvider) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewIdentityProviderID()
	}

	return nil
}

func (a *IdentityProvider) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &IdentityProvider{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

// ValidateConfig validates the Config field based on ProviderType.
func (ip *IdentityProvider) ValidateConfig() error {
	if len(ip.Config) == 0 {
		return fmt.Errorf("config is required")
	}

	switch ip.ProviderType {
	case ProviderTypeOIDC:
		cfg, err := ip.GetOpenIDConfig()
		if err != nil {
			return fmt.Errorf("invalid openid config: %w", err)
		}
		return cfg.Validate()

	case ProviderTypeGoogle:
		cfg, err := ip.GetGoogleConfig()
		if err != nil {
			return fmt.Errorf("invalid google config: %w", err)
		}
		return cfg.Validate()

	case ProviderTypeGitHub:
		cfg, err := ip.GetGitHubConfig()
		if err != nil {
			return fmt.Errorf("invalid github config: %w", err)
		}
		return cfg.Validate()

	default:
		return fmt.Errorf("unknown provider type: %s", ip.ProviderType)
	}
}

// GetClientID returns the client_id from the provider config, regardless of type.
func (ip *IdentityProvider) GetClientID() (string, error) {
	if len(ip.Config) == 0 {
		return "", nil
	}

	var cfg providers.BaseConfig
	if err := json.Unmarshal(ip.Config, &cfg); err != nil {
		return "", err
	}
	return cfg.ClientID, nil
}

// GetOpenIDConfig unmarshals the Config as OpenIDConfig.
func (ip *IdentityProvider) GetOpenIDConfig() (*providers.OpenIDConfig, error) {
	var cfg providers.OpenIDConfig
	if err := json.Unmarshal(ip.Config, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SetOpenIDConfig marshals and sets the Config from OpenIDConfig.
func (ip *IdentityProvider) SetOpenIDConfig(cfg *providers.OpenIDConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	ip.Config = data
	ip.ProviderType = ProviderTypeOIDC
	return nil
}

// GetGoogleConfig unmarshals the Config as GoogleConfig.
func (ip *IdentityProvider) GetGoogleConfig() (*providers.GoogleConfig, error) {
	var cfg providers.GoogleConfig
	if err := json.Unmarshal(ip.Config, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SetGoogleConfig marshals and sets the Config from GoogleConfig.
func (ip *IdentityProvider) SetGoogleConfig(cfg *providers.GoogleConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	ip.Config = data
	ip.ProviderType = ProviderTypeGoogle
	return nil
}

// GetGitHubConfig unmarshals the Config as GitHubConfig.
func (ip *IdentityProvider) GetGitHubConfig() (*providers.GitHubConfig, error) {
	var cfg providers.GitHubConfig
	if err := json.Unmarshal(ip.Config, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SetGitHubConfig marshals and sets the Config from GitHubConfig.
func (ip *IdentityProvider) SetGitHubConfig(cfg *providers.GitHubConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	ip.Config = data
	ip.ProviderType = ProviderTypeGitHub
	return nil
}
