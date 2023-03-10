package models

import (
	"time"
)

type Org struct {
	Model
	CreatedByID string
	Name        string `gorm:"uniqueIndex"`
	Apps        []App  `faker:"-"`
	IsNew       bool   `gorm:"-:all"`
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
