package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type TemporalPayload struct {
	ID        string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Contents []byte `json:"contents,omitzero" gorm:"type:jsonb" swaggertype:"string" features:"template" temporaljson:"contents,omitzero,omitempty"`
}

func (a *TemporalPayload) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewTemporalPayload()
	}
	return nil
}
