package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type OAuthClientSecret struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	ClientID   string     `json:"client_id,omitzero" gorm:"notnull" temporaljson:"client_id,omitzero,omitempty"`
	SecretHash string     `json:"-" gorm:"notnull" temporaljson:"-"`
	Hint       string     `json:"hint,omitzero" temporaljson:"hint,omitzero,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitzero" temporaljson:"expires_at,omitzero,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitzero" temporaljson:"last_used_at,omitzero,omitempty"`
}

func (a *OAuthClientSecret) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:        indexes.Name(db, &OAuthClientSecret{}, "secret_hash"),
			Columns:     []string{"secret_hash"},
			UniqueValue: generics.NewNullBool(true),
		},
		{
			Name:    indexes.Name(db, &OAuthClientSecret{}, "client_id"),
			Columns: []string{"client_id"},
			Option:  "WHERE deleted_at = 0",
		},
	}
}

func (a *OAuthClientSecret) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewOAuthClientSecretID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *OAuthClientSecret) Active(now time.Time) bool {
	return a.ExpiresAt == nil || now.Before(*a.ExpiresAt)
}
