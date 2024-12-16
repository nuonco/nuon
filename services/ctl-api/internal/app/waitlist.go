package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type Waitlist struct {
	ID          string  `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time `json:"updated_at" gorm:"notnull"`

	OrgName string `json:"org_name" gorm:"not null;default:null"`
}

func (c *Waitlist) BeforeSave(tx *gorm.DB) error {
	c.ID = domains.NewWaitListID()

	return nil
}
