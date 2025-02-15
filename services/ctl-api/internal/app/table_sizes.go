package app

import "gorm.io/gorm"

type TableSize struct {
	TableName  string  `json"table_name" gorm:"->;-:migration"`
	SizePretty string  `json"size_pretty" gorm:"->;-:migration"`
	SizeBytes  float64 `json:"size_bytes" gorm:"->;-:migration"`
}

func (*TableSize) UseView() bool {
	return true
}

func (*TableSize) ViewVersion() string {
	return "v1"
}

func (m TableSize) GetTableOptions() (string, bool) {
	return "", false
}

func (r TableSize) MigrateDB(db *gorm.DB) *gorm.DB {
	return db
}
