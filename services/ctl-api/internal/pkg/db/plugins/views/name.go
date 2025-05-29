package views

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

// TableOrViewName returns the table or view name for an object, appending the provided string.
func TableOrViewName(db *gorm.DB, obj ViewModel, appendStr string) string {
	tableName := plugins.TableName(db, obj)
	disableViewTableName := fmt.Sprintf("%s%s", tableName, appendStr)

	if !obj.UseView() {
		fmt.Println("rb - disable views: ", disableViewTableName)
		return disableViewTableName
	}

	disable, ok := db.InstanceGet(DisableViewsKey)
	fmt.Println("rb - disable views: ", disable)
	if ok && disable.(bool) {
		return disableViewTableName
	}

	fmt.Println("rb - enable views: ", tableName, appendStr)
	return fmt.Sprintf("%s_view_%s%s", tableName, obj.ViewVersion(), appendStr)
}

// DefaultTableName returns the default table name for an object, appending the provided string.
// This should be used when scopes.WithDisableViews is applied to the query.
func DefaultTableName(db *gorm.DB, obj any, appendStr string) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s%s", tableName, appendStr)
}

func DefaultViewName(db *gorm.DB, obj any, version int) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s_view_v%d", tableName, version)
}

func CustomViewName(db *gorm.DB, obj any, name string) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s_%s", tableName, name)
}
