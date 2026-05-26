package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type TemporalBlob struct {
	ID        string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	S3Key    string `json:"s3_key" gorm:"type:text;not null" temporaljson:"s3_key,omitzero,omitempty"`
	Checksum string `json:"checksum" gorm:"type:text" temporaljson:"checksum,omitzero,omitempty"`
	Size     int64  `json:"size" temporaljson:"size,omitzero,omitempty"`
}

func (a *TemporalBlob) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewTemporalBlob()
	}
	return nil
}
