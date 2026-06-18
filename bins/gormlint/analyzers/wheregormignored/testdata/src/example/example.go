package example

import "gorm.io/gorm"

type User struct {
	ID       string
	Name     string
	OrgID    string
	Computed string `gorm:"-"`
	Links    string `gorm:"-"`
}

func badWhereGormIgnored(db *gorm.DB) {
	db.Where(User{Computed: "val"}).First(&User{}) // want `field "Computed" has gorm:"-" tag and will be silently ignored in this Where clause`
}

func badWhereGormIgnoredPointer(db *gorm.DB) {
	db.Where(&User{Links: "val"}).First(&User{}) // want `field "Links" has gorm:"-" tag and will be silently ignored in this Where clause`
}

func badWhereGormIgnoredMixed(db *gorm.DB) {
	db.Where(User{Name: "good", Computed: "bad"}).First(&User{}) // want `field "Computed" has gorm:"-" tag and will be silently ignored in this Where clause`
}

func goodWhereNormalFields(db *gorm.DB) {
	db.Where(User{Name: "val", OrgID: "org123"}).First(&User{})
}

func goodWhereStringCondition(db *gorm.DB) {
	db.Where("name = ?", "val").First(&User{})
}
