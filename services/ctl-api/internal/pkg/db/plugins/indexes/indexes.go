package indexes

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

const (
	defaultPrefix string = "idx"
	maxLen        int    = 63
)

func Name(db *gorm.DB, obj any, name string) string {
	tableName := plugins.TableName(db, obj)
	idxName := fmt.Sprintf("%s_%s_%s", defaultPrefix, tableName, name)

	if len(idxName) > 63 {
		panic(fmt.Sprintf("index %s for table %s is too long", idxName, tableName))
	}

	return idxName
}
