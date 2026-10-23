package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type DeviceAuthorizationStatus string

const (
	DeviceAuthorizationStatusPending  DeviceAuthorizationStatus = "pending"
	DeviceAuthorizationStatusApproved DeviceAuthorizationStatus = "approved"
	DeviceAuthorizationStatusDenied   DeviceAuthorizationStatus = "denied"
	DeviceAuthorizationStatusConsumed DeviceAuthorizationStatus = "consumed"
)

// DeviceAuthorization backs the OAuth 2.0 device authorization grant (RFC 8628). The server
// issues both codes: DeviceCode is the high-entropy secret the CLI polls with, UserCode is the
// short code the user types in at the verification page.
type DeviceAuthorization struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	DeviceCode string `gorm:"unique;not null" json:"-" temporaljson:"device_code,omitzero,omitempty"`
	UserCode   string `gorm:"unique;not null" json:"-" temporaljson:"user_code,omitzero,omitempty"`

	RequesterIP        string `json:"requester_ip,omitzero" temporaljson:"requester_ip,omitzero,omitempty"`
	RequesterUserAgent string `json:"requester_user_agent,omitzero" temporaljson:"requester_user_agent,omitzero,omitempty"`

	Status    DeviceAuthorizationStatus `gorm:"not null;default:pending" json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	ExpiresAt time.Time                 `gorm:"not null" json:"expires_at,omitzero" temporaljson:"expires_at,omitzero,omitempty"`

	// AccountID is set on approval. No foreign key: the row is created before anyone has
	// approved it, when account_id is still empty.
	AccountID     string `gorm:"index" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	SourceTokenID string `gorm:"index" json:"-" temporaljson:"source_token_id,omitzero,omitempty"`
}

func (a *DeviceAuthorization) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewDeviceAuthorizationID()
	}
	return nil
}
