package views

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

func TableOrViewName(db *gorm.DB, obj ViewModel, appendStr string) string {
	tableName := plugins.TableName(db, obj)

	if !obj.UseView() {
		return fmt.Sprintf("%s%s", tableName, appendStr)
	}

	return fmt.Sprintf("%s_view_%s%s", tableName, obj.ViewVersion(), appendStr)
}

func DefaultViewName(db *gorm.DB, obj any, version int) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s_view_v%d", tableName, version)
}

func CustomViewName(db *gorm.DB, obj any, name string) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s_%s", tableName, name)
}
