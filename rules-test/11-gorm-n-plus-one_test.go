package main

import (
	"testing"

	"gorm.io/gorm"
)

type NPlusOneUser struct {
	ID   uint
	Name string
}

type NPlusOneOrder struct {
	ID     uint
	UserID uint
	Amount float64
}

// BadUsage_FirstInLoop should trigger the rule
func BadUsage_FirstInLoop(t *testing.T) {
	var db *gorm.DB
	var users []NPlusOneUser
	var order NPlusOneOrder

	// BUG: N+1 query - calling First for each user
	for _, user := range users {
		db.First(&order, user.ID)
	}
	_ = order
}

// BadUsage_FindInLoop should trigger the rule
func BadUsage_FindInLoop(t *testing.T) {
	var db *gorm.DB
	var userIDs []uint
	var orders []NPlusOneOrder

	// BUG: N+1 query - calling Find for each userID
	for _, id := range userIDs {
		db.Where("user_id = ?", id).Find(&orders)
	}
	_ = orders
}

// BadUsage_CreateInLoop should trigger the rule
func BadUsage_CreateInLoop(t *testing.T) {
	var db *gorm.DB
	var users []NPlusOneUser

	// BUG: N+1 inserts - should use CreateInBatches instead
	for _, user := range users {
		db.Create(&user)
	}
}

// BadUsage_UpdateInLoop should trigger the rule
func BadUsage_UpdateInLoop(t *testing.T) {
	var db *gorm.DB
	var userIDs []uint

	// BUG: N+1 updates - should use batch update
	for _, id := range userIDs {
		db.Model(&NPlusOneUser{}).Where("id = ?", id).Update("name", "updated")
	}
}

// BadUsage_DeleteInLoop should trigger the rule
func BadUsage_DeleteInLoop(t *testing.T) {
	var db *gorm.DB
	var userIDs []uint

	// BUG: N+1 deletes - should use WHERE IN
	for _, id := range userIDs {
		db.Delete(&NPlusOneUser{}, id)
	}
}

// BadUsage_SaveInLoop should trigger the rule
func BadUsage_SaveInLoop(t *testing.T) {
	var db *gorm.DB
	var users []NPlusOneUser

	// BUG: N+1 saves
	for i := range users {
		users[i].Name = "updated"
		db.Save(&users[i])
	}
}

// BadUsage_CountInLoop should trigger the rule
func BadUsage_CountInLoop(t *testing.T) {
	var db *gorm.DB
	var userIDs []uint
	var count int64

	// BUG: N+1 count queries
	for _, id := range userIDs {
		db.Model(&NPlusOneOrder{}).Where("user_id = ?", id).Count(&count)
	}
	_ = count
}

// BadUsage_PreloadInLoop should trigger the rule
func BadUsage_PreloadInLoop(t *testing.T) {
	var db *gorm.DB
	var userIDs []uint
	var user NPlusOneUser

	// BUG: N+1 preload queries
	for _, id := range userIDs {
		db.Preload("Orders").First(&user, id)
	}
	_ = user
}

// BadUsage_CStyleForLoop should trigger the rule
func BadUsage_CStyleForLoop(t *testing.T) {
	var db *gorm.DB
	userIDs := []uint{1, 2, 3}
	var order NPlusOneOrder

	// BUG: N+1 in C-style for loop
	for i := 0; i < len(userIDs); i++ {
		db.First(&order, userIDs[i])
	}
	_ = order
}

// GoodUsage_BatchQueryWithWhereIn should NOT trigger the rule
func GoodUsage_BatchQueryWithWhereIn(t *testing.T) {
	var db *gorm.DB
	userIDs := []uint{1, 2, 3}
	var orders []NPlusOneOrder

	// OK: Single query with WHERE IN
	db.Where("user_id IN ?", userIDs).Find(&orders)
	_ = orders
}

// GoodUsage_CreateInBatches should NOT trigger the rule
func GoodUsage_CreateInBatches(t *testing.T) {
	var db *gorm.DB
	users := []NPlusOneUser{{Name: "a"}, {Name: "b"}}

	// OK: Batch insert
	db.CreateInBatches(users, 100)
}

// GoodUsage_BatchUpdate should NOT trigger the rule
func GoodUsage_BatchUpdate(t *testing.T) {
	var db *gorm.DB
	userIDs := []uint{1, 2, 3}

	// OK: Single batch update
	db.Model(&NPlusOneUser{}).Where("id IN ?", userIDs).Update("name", "updated")
}

// GoodUsage_BatchDelete should NOT trigger the rule
func GoodUsage_BatchDelete(t *testing.T) {
	var db *gorm.DB
	userIDs := []uint{1, 2, 3}

	// OK: Single batch delete
	db.Where("id IN ?", userIDs).Delete(&NPlusOneUser{})
}

// GoodUsage_PreloadWithFind should NOT trigger the rule
func GoodUsage_PreloadWithFind(t *testing.T) {
	var db *gorm.DB
	var users []NPlusOneUser

	// OK: Preload loads all related records in single query
	db.Preload("Orders").Find(&users)
	_ = users
}

// GoodUsage_JoinsQuery should NOT trigger the rule
func GoodUsage_JoinsQuery(t *testing.T) {
	var db *gorm.DB
	var users []NPlusOneUser

	// OK: Single query with join
	db.Joins("Orders").Find(&users)
	_ = users
}
