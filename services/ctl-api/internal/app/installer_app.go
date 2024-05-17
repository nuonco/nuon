package app

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type InstallerApp struct {
	DeletedAt soft_delete.DeletedAt `json:"-"`

	AppID string `gorm:"primary_key"`
	App   App

	InstallerID string `gorm:"primary_key"`
	Installer   Installer
}

func (c *InstallerApp) BeforeSave(tx *gorm.DB) error {
	return nil
}
