package example

import "gorm.io/gorm"

type User struct {
	ID   string
	Name string
}

func badFindNoLimit(db *gorm.DB) {
	var users []User
	db.Where(User{Name: "alice"}).Find(&users) // want `Find\(\) without Limit\(\) in query chain`
	_ = users
}

func badScanNoLimit(db *gorm.DB) {
	var users []User
	db.Where(User{Name: "alice"}).Scan(&users) // want `Scan\(\) without Limit\(\) in query chain`
	_ = users
}

func badFindWithOrderNoLimit(db *gorm.DB) {
	var users []User
	db.Where(User{Name: "alice"}).Order("name").Find(&users) // want `Find\(\) without Limit\(\) in query chain`
	_ = users
}

func goodFindWithLimit(db *gorm.DB) {
	var users []User
	db.Where(User{Name: "alice"}).Limit(100).Find(&users)
	_ = users
}

func goodFindWithInlineCondition(db *gorm.DB) {
	var user User
	db.Find(&user, "id = ?", "123")
}

func goodFirst(db *gorm.DB) {
	var user User
	db.Where(User{Name: "alice"}).First(&user)
}
