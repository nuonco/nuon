// install.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Install struct {
	Model
	CreatedByID uuid.UUID

	Name  string
	AppID uuid.UUID
	App   App

	Domain   Domain          // all the domain stuff
	Settings InstallSettings `gorm:"-" faker:"-"`

	AWSSettings *AWSSettings `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" faker:"-"`
	GCPSettings *GCPSettings `faker:"-"`
}

func (Install) IsNode() {}

func (i Install) GetID() string {
	return i.Model.ID.String()
}

func (i Install) GetCreatedAt() time.Time {
	return i.Model.CreatedAt
}

func (i Install) GetUpdatedAt() time.Time {
	return i.Model.UpdatedAt
}
