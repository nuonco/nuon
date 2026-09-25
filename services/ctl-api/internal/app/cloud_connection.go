package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type CloudConnectionStatus string

const (
	CloudConnectionStatusPending  CloudConnectionStatus = "pending"
	CloudConnectionStatusVerified CloudConnectionStatus = "verified"
	CloudConnectionStatusError    CloudConnectionStatus = "error"
)

type CloudConnectionAuthMode string

const (
	CloudConnectionAuthModeOIDC   CloudConnectionAuthMode = "oidc"
	CloudConnectionAuthModeLegacy CloudConnectionAuthMode = "legacy"
)

type CloudConnectionCapability string

const (
	CloudConnectionCapabilityStacks CloudConnectionCapability = "stacks"
	CloudConnectionCapabilityImages CloudConnectionCapability = "images"
)

type CloudConnection struct {
	ID                    string                      `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitempty"`
	CreatedByID           string                      `gorm:"not null;default:null" json:"created_by_id" temporaljson:"created_by_id,omitempty"`
	CreatedBy             Account                     `json:"-" temporaljson:"created_by,omitempty"`
	CreatedAt             time.Time                   `gorm:"notnull" json:"created_at" temporaljson:"created_at,omitempty"`
	UpdatedAt             time.Time                   `gorm:"notnull" json:"updated_at" temporaljson:"updated_at,omitempty"`
	DeletedAt             soft_delete.DeletedAt       `gorm:"uniqueIndex:idx_cloud_connections_org_platform_principal_deleted" json:"-" temporaljson:"deleted_at,omitempty"`
	OrgID                 string                      `gorm:"notnull;uniqueIndex:idx_cloud_connections_org_platform_principal_deleted" json:"org_id" temporaljson:"org_id,omitempty"`
	Org                   Org                         `json:"-" temporaljson:"org,omitempty"`
	Name                  string                      `gorm:"notnull" json:"name" temporaljson:"name,omitempty"`
	Platform              CloudPlatform               `gorm:"notnull;uniqueIndex:idx_cloud_connections_org_platform_principal_deleted" json:"platform" temporaljson:"platform,omitempty"`
	TargetID              string                      `gorm:"notnull" json:"target_id" temporaljson:"target_id,omitempty"`
	Principal             string                      `gorm:"notnull;uniqueIndex:idx_cloud_connections_org_platform_principal_deleted" json:"principal" temporaljson:"principal,omitempty"`
	TenantID              string                      `gorm:"notnull;default:''" json:"tenant_id,omitempty" temporaljson:"tenant_id,omitempty"`
	IdentityProvider      string                      `gorm:"notnull;default:''" json:"identity_provider,omitempty" temporaljson:"identity_provider,omitempty"`
	DefaultRegion         string                      `gorm:"notnull;default:''" json:"default_region,omitempty" temporaljson:"default_region,omitempty"`
	AuthMode              CloudConnectionAuthMode     `gorm:"notnull;default:''" json:"auth_mode,omitempty" temporaljson:"auth_mode,omitempty"`
	Status                CloudConnectionStatus       `gorm:"notnull;default:'pending'" json:"status" temporaljson:"status,omitempty"`
	StatusMessage         string                      `gorm:"notnull;default:''" json:"status_message,omitempty" temporaljson:"status_message,omitempty"`
	LastVerifiedAt        *time.Time                  `json:"last_verified_at,omitempty" temporaljson:"last_verified_at,omitempty"`
	RequestedCapabilities []CloudConnectionCapability `gorm:"type:jsonb;serializer:json;notnull;default:'[]'" json:"requested_capabilities" temporaljson:"requested_capabilities,omitempty"`
	Capabilities          []CloudConnectionCapability `gorm:"type:jsonb;serializer:json;notnull;default:'[]'" json:"capabilities" temporaljson:"capabilities,omitempty"`
	Registries            []string                    `gorm:"type:jsonb;serializer:json" json:"registries,omitempty" temporaljson:"registries,omitempty"`
}

func (c *CloudConnection) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = domains.NewCloudConnectionID()
	}
	if c.Status == "" {
		c.Status = CloudConnectionStatusPending
	}
	if c.RequestedCapabilities == nil {
		c.RequestedCapabilities = append([]CloudConnectionCapability(nil), c.Capabilities...)
	}
	if c.Capabilities == nil {
		c.Capabilities = []CloudConnectionCapability{}
	}
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (c *CloudConnection) HasCapability(capability CloudConnectionCapability) bool {
	for _, value := range c.Capabilities {
		if value == capability {
			return true
		}
	}
	return false
}

func (c *CloudConnection) HasRequestedCapability(capability CloudConnectionCapability) bool {
	for _, value := range c.RequestedCapabilities {
		if value == capability {
			return true
		}
	}
	return false
}
