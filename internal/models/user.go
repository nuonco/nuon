// user.go
package models

import "time"

type User struct {
	Model

	Email      string `gorm:"unique"`
	ExternalID string
	FirstName  string
	LastName   string
	IsAdmin    bool
	IsNew      bool `gorm:"-:all"`

	Orgs []Org `gorm:"many2many:user_orgs" fake:"skip"`
}

func (User) IsNode() {}

func (u User) GetID() string {
	return u.Model.ID.String()
}

func (u User) GetCreatedAt() time.Time {
	return u.Model.CreatedAt
}

func (u User) GetUpdatedAt() time.Time {
	return u.Model.UpdatedAt
}
