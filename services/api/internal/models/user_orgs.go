package models

type UserOrg struct {
	ModelV2

	UserID string
	OrgID  string `gorm:"primaryKey"`
	IsNew  bool   `gorm:"-:all"`
}

func (UserOrg) IsNode() {}
