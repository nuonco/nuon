package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type TokenType string

const (
	TokenTypeAuth0       TokenType = "auth0"
	TokenTypeIntegration TokenType = "integration"
	TokenTypeCanary      TokenType = "canary"
	TokenTypeAdmin       TokenType = "admin"
)

type UserToken struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	Token     string    `gorm:"uniqueIndex;notnull" json:"-"`
	TokenType TokenType `json:"token_type"`

	// claim data
	Subject   string    `json:"subject" gorm:"notnull"`
	ExpiresAt time.Time `json:"expires_at" gorm:"notnull"`
	IssuedAt  time.Time `json:"issued_at" gorm:"notnull"`
	Issuer    string    `json:"issuer" gorm:"notnull"`
	Email     string    `json:"email"`
}

func (u *UserToken) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserTokenID()
	return nil
}
