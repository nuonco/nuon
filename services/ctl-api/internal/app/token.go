package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type TokenType string

const (
	TokenTypeAuth0       TokenType = "auth0"
	TokenTypeAdmin       TokenType = "admin"
	TokenTypeStatic      TokenType = "static"
	TokenTypeIntegration TokenType = "integration"
	TokenTypeCanary      TokenType = "canary"
	TokenTypeNuon        TokenType = "nuon"
)

type Token struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	AccountID string `json:"account_id" temporaljson:"account_id,omitzero,omitempty"`

	Token     string    `gorm:"unique" json:"-" temporaljson:"token,omitzero,omitempty"`
	TokenType TokenType `json:"token_type" temporaljson:"token_type,omitzero,omitempty"`

	// claim data
	ExpiresAt time.Time `json:"expires_at" gorm:"notnull" temporaljson:"expires_at,omitzero,omitempty"`
	IssuedAt  time.Time `json:"issued_at" gorm:"notnull" temporaljson:"issued_at,omitzero,omitempty"`
	Issuer    string    `json:"issuer" gorm:"notnull;default null" temporaljson:"issuer,omitzero,omitempty"`
}

func (a *Token) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewUserTokenID()
	return nil
}
