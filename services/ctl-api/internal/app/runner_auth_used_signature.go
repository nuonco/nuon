package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerAuthUsedSignature struct {
	ID        string    `gorm:"primarykey" json:"id,omitzero"`
	CreatedAt time.Time `json:"created_at,omitzero"`

	RunnerID      string `gorm:"not null" json:"runner_id,omitzero"`
	SignatureHash string `gorm:"not null" json:"signature_hash,omitzero"`

	ExpiresAt time.Time `gorm:"not null" json:"expires_at,omitzero"`
}

func (s *RunnerAuthUsedSignature) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = domains.NewRunnerAuthUsedSignatureID()
	}

	return nil
}

func (s *RunnerAuthUsedSignature) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:        indexes.Name(db, s, "signature_hash"),
			Columns:     []string{"signature_hash"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name:    indexes.Name(db, s, "expires_at"),
			Columns: []string{"expires_at"},
		},
	}
}
