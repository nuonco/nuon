package app

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type PSQLTableSize struct {
	TableName  string  `json"table_name" gorm:"->;-:migration"`
	SizePretty string  `json"size_pretty" gorm:"->;-:migration"`
	SizeBytes  float64 `json:"size_bytes" gorm:"->;-:migration"`
}

func (*PSQLTableSize) UseView() bool {
	return true
}

func (*PSQLTableSize) ViewVersion() string {
	return "v1"
}

func (i *PSQLTableSize) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &PSQLTableSize{}, 1),
			SQL:  viewsql.PSQLTableSizesV1,
		},
	}
}

func (m PSQLTableSize) GetTableOptions() (string, bool) {
	return "", false
}

func (r PSQLTableSize) MigrateDB(db *gorm.DB) *gorm.DB {
	return db
}
