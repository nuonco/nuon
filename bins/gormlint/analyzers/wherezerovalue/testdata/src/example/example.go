package example

import "gorm.io/gorm"

type User struct {
	ID     string
	Name   string
	Age    int
	Active bool
	Score  float64
}

func badZeroInt(db *gorm.DB) {
	db.Where(User{Age: 0}).First(&User{}) // want `field "Age" has zero value and will be silently ignored by GORM in this Where clause`
}

func badZeroBool(db *gorm.DB) {
	db.Where(User{Active: false}).First(&User{}) // want `field "Active" has zero value and will be silently ignored by GORM in this Where clause`
}

func badZeroString(db *gorm.DB) {
	db.Where(User{Name: ""}).First(&User{}) // want `field "Name" has zero value and will be silently ignored by GORM in this Where clause`
}

func badZeroFloat(db *gorm.DB) {
	db.Where(User{Score: 0.0}).First(&User{}) // want `field "Score" has zero value and will be silently ignored by GORM in this Where clause`
}

func goodNonZeroValues(db *gorm.DB) {
	db.Where(User{Name: "alice", Age: 25, Active: true}).First(&User{})
}

func goodVariableValue(db *gorm.DB) {
	age := 0
	db.Where(User{Age: age}).First(&User{})
}
