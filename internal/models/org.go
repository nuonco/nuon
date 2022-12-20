package models

import (
	"time"

	"github.com/google/uuid"
)

type Org struct {
	Model
	CreatedByID uuid.UUID `gorm:"type:uuid" json:"owner_id"`

	Slug       string `gorm:"uniqueIndex"`
	Name       string `gorm:"uniqueIndex"`
	Users      []User `gorm:"many2many:user_orgs" fake:"skip" json:"-"`
	Apps       []App  `fake:"skip" json:"-"`
	WorkflowID string `json:"-"`
	IsNew      bool   `gorm:"-:all"`
}

func (Org) IsNode() {}

func (o Org) GetID() string {
	return o.Model.ID.String()
}

func (o Org) GetCreatedAt() time.Time {
	return o.Model.CreatedAt
}

func (o Org) GetUpdatedAt() time.Time {
	return o.Model.UpdatedAt
}
