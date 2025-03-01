package app

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type CHTableSize struct {
	TableName  string  `json"table_name" gorm:"->;-:migration"`
	SizePretty string  `json"size_pretty" gorm:"->;-:migration"`
	SizeBytes  float64 `json:"size_bytes" gorm:"->;-:migration"`
}

func (*CHTableSize) UseView() bool {
	return true
}

func (*CHTableSize) ViewVersion() string {
	return "v1"
}

func (i *CHTableSize) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &CHTableSize{}, 1),
			SQL:  viewsql.CHTableSizesV1,
		},
	}
}

func (m CHTableSize) GetTableOptions() (string, bool) {
	return "", false
}
