package example

import "gorm.io/gorm"

type User struct {
	ID   string
	Name string
}

func badJoinsRawSQL(db *gorm.DB) {
	db.Joins("JOIN orders ON orders.user_id = users.id").Find(&User{}) // want `raw SQL string in Joins\(\)`
}

func badOrderRawSQL(db *gorm.DB) {
	db.Order("created_at DESC").Find(&User{}) // want `raw SQL string in Order\(\)`
}

func badSelectRawSQL(db *gorm.DB) {
	db.Select("id, name").Find(&User{}) // want `raw SQL string in Select\(\)`
}

func badGroupRawSQL(db *gorm.DB) {
	db.Group("org_id").Find(&User{}) // want `raw SQL string in Group\(\)`
}

func badHavingRawSQL(db *gorm.DB) {
	db.Having("count(*) > ?", 5).Find(&User{}) // want `raw SQL string in Having\(\)`
}

func goodJoinsAssociation(db *gorm.DB) {
	db.Joins("Orders").Find(&User{})
}
