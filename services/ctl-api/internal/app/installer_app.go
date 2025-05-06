package app

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type InstallerApp struct {
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" gorm:"primary_key" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `json:"app,omitzero" temporaljson:"app,omitzero,omitempty"`

	InstallerID string    `json:"installer_id,omitzero" gorm:"primary_key" temporaljson:"installer_id,omitzero,omitempty"`
	Installer   Installer `json:"installer,omitzero" temporaljson:"installer,omitzero,omitempty"`
}

func (c *InstallerApp) BeforeSave(tx *gorm.DB) error {
	return nil
}
