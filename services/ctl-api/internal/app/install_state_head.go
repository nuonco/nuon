package app

import (
	"github.com/nuonco/nuon/pkg/generics"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

type InstallStateHead struct {
	InstallID      string                 `gorm:"primaryKey" json:"-"`
	Install        Install                `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	OrgID          string                 `gorm:"not null" json:"-"`
	Revision       int64                  `gorm:"not null;default:1" json:"-"`
	CurrentStateID string                 `gorm:"not null;default:''" json:"-"`
	StaleAt        generics.NullTime      `gorm:"type:timestamptz;default:null" json:"-"`
	StalePartials  []pkgstate.PartialName `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"-"`
}
