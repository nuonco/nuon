package models

import (
	"github.com/google/uuid"
)

type UserOrg struct {
	Model

	UserID string
	OrgID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	IsNew  bool      `gorm:"-:all"`
}

func (UserOrg) IsNode() {}
