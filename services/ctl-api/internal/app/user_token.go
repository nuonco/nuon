package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type UserToken struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Token string `gorm:"uniqueIndex" json:"-"`

	// claim data
	Subject   string    `json:"subject"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Issuer    string    `json:"issuer"`
}

func (u *UserToken) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserTokenID()
	return nil
}
