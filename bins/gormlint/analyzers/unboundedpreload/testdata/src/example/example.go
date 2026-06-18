package example

import "gorm.io/gorm"

type User struct {
	ID     string
	Orders []Order
}

type Order struct {
	ID     string
	UserID string
}

func badNoScopingFunc(db *gorm.DB) {
	db.Preload("Orders").Find(&User{}) // want `unbounded Preload without scoping function`
}

func badScopingFuncNoLimit(db *gorm.DB) {
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB { // want `Preload scoping function has no Limit`
		return db.Order("created_at DESC")
	}).Find(&User{})
}

func goodScopingFuncWithLimit(db *gorm.DB) {
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(10)
	}).Find(&User{})
}

func goodScopingFuncWithWhereAndLimit(db *gorm.DB) {
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", "active").Limit(5)
	}).Find(&User{})
}
