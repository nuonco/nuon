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
)

type Token struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	AccountID string `json:"account_id"`

	Token     string    `gorm:"uniqueIndex;notnull" json:"-"`
	TokenType TokenType `json:"token_type"`

	// Deprecated
	Email   string
	Subject string

	// claim data
	ExpiresAt time.Time `json:"expires_at" gorm:"notnull"`
	IssuedAt  time.Time `json:"issued_at" gorm:"notnull"`
	Issuer    string    `json:"issuer" gorm:"notnull;default null"`
}

func (a *Token) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewUserTokenID()
	return nil
}
