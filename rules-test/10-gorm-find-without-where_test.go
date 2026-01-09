package main

import (
	"testing"

	"gorm.io/gorm"
)

// BadUsage_FindWithoutWhere should trigger the rule
func BadUsage_FindWithoutWhere(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Find without WHERE loads the entire table
	db.Find(&users)
}

// BadUsage_FindWithoutWhereModel should trigger the rule
func BadUsage_FindWithoutWhereModel(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Model().Find() without WHERE still loads all records
	db.Model(&User{}).Find(&users)
}

// BadUsage_FindWithoutWherePreload should trigger the rule
func BadUsage_FindWithoutWherePreload(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Preload doesn't filter the main query
	db.Preload("Orders").Find(&users)
}

// BadUsage_FindWithoutWhereOrder should trigger the rule
func BadUsage_FindWithoutWhereOrder(t *testing.T) {
	var db *gorm.DB
	var users []User

	// BUG: Order doesn't limit the result set
	db.Order("created_at DESC").Find(&users)
}

// GoodUsage_FindWithWhere should NOT trigger the rule
func GoodUsage_FindWithWhere(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: WHERE clause filters the results
	db.Where("active = ?", true).Find(&users)
	_ = users
}

// GoodUsage_FindWithWhereChained should NOT trigger the rule
func GoodUsage_FindWithWhereChained(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: WHERE with other clauses
	db.Where("role = ?", "admin").Order("name").Find(&users)
	_ = users
}

// GoodUsage_FindWithLimit should NOT trigger the rule
func GoodUsage_FindWithLimit(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Limit prevents loading all records
	db.Limit(100).Find(&users)
	_ = users
}

// GoodUsage_FindWithInlineCondition should NOT trigger the rule
func GoodUsage_FindWithInlineCondition(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Inline condition in Find
	db.Find(&users, "active = ?", true)
	_ = users
}

// GoodUsage_FindWithWhereAndPreload should NOT trigger the rule
func GoodUsage_FindWithWhereAndPreload(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: WHERE clause present
	db.Where("active = ?", true).Preload("Orders").Find(&users)
	_ = users
}

// GoodUsage_FindWithLimitAndOrder should NOT trigger the rule
func GoodUsage_FindWithLimitAndOrder(t *testing.T) {
	var db *gorm.DB
	var users []User

	// OK: Limit prevents unbounded query
	db.Order("created_at DESC").Limit(50).Find(&users)
	_ = users
}
