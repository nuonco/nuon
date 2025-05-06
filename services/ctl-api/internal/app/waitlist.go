package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type Waitlist struct {
	ID          string  `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`

	OrgName string `json:"org_name,omitzero" gorm:"not null;default:null" temporaljson:"org_name,omitzero,omitempty"`
}

func (c *Waitlist) BeforeSave(tx *gorm.DB) error {
	c.ID = domains.NewWaitListID()

	return nil
}
