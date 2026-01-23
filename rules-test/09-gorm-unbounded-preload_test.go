package main

import (
	"testing"

	"gorm.io/gorm"
)

type User struct {
	ID     uint
	Name   string
	Orders []Order
}

type Order struct {
	ID     uint
	UserID uint
	Amount float64
}

// BadUsage_UnboundedPreload should trigger the rule
func BadUsage_UnboundedPreload(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Preload without scoping function loads ALL orders
	db.Preload("Orders").Find(&users)
}

// BadUsage_ChainedUnboundedPreload should trigger the rule
func BadUsage_ChainedUnboundedPreload(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Both Preloads are unbounded
	db.Preload("Orders").Preload("Profile").Find(&users)
}

// BadUsage_UnboundedPreloadWithWhere should trigger the rule
func BadUsage_UnboundedPreloadWithWhere(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Even with Where, the Preload is still unbounded
	db.Where("active = ?", true).Preload("Orders").Find(&users)
}

// GoodUsage_PreloadWithScopingFunction should NOT trigger the rule
func GoodUsage_PreloadWithScopingFunction(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Using scoping function to limit preloaded records
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Limit(10)
	}).Find(&users)
	_ = users
}

// GoodUsage_PreloadWithOrderAndLimit should NOT trigger the rule
func GoodUsage_PreloadWithOrderAndLimit(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Using scoping function with Order and Limit
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(5)
	}).Find(&users)
	_ = users
}

// GoodUsage_PreloadWithConditionFunc should NOT trigger the rule
func GoodUsage_PreloadWithConditionFunc(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Scoping function with Where clause
	db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", "completed").Limit(10)
	}).Find(&users)
	_ = users
}
