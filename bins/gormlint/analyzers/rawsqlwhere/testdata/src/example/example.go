package example

import "gorm.io/gorm"

type User struct {
	ID    string
	Name  string
	OrgID string
}

func badRawSQLWhere(db *gorm.DB) {
	db.Where("name = ?", "alice").First(&User{}) // want `raw SQL string in Where clause`
}

func badRawSQLWhereMultiple(db *gorm.DB) {
	db.Where("name = ? AND org_id = ?", "alice", "org1").First(&User{}) // want `raw SQL string in Where clause`
}

func goodStructWhere(db *gorm.DB) {
	db.Where(User{Name: "alice"}).First(&User{})
}

func goodStructWherePointer(db *gorm.DB) {
	db.Where(&User{Name: "alice"}).First(&User{})
}
