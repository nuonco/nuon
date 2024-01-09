package db

import (
	"gorm.io/gorm"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	afterQueryPlug := &afterQueryPlugin{}
	db.Callback().
		Query().
		After("gorm:query").
		Register("nuon:after_query", afterQueryPlug.plugin)

	return nil
}
