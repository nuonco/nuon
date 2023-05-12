package models

type UserOrg struct {
	Model

	UserID string
	OrgID  string `gorm:"primaryKey"`
	IsNew  bool   `gorm:"-:all"`
}

func (UserOrg) IsNode() {}
