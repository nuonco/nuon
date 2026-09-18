package views

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func RewriteName(db *gorm.DB, name string) string {
	return pluginFromDB(db).rewriteName(name)
}

// TableOrViewName returns the table or view name for an object, appending the provided string.
func TableOrViewName(db *gorm.DB, obj ViewModel, appendStr string) string {
	tableName := plugins.TableName(db, obj)
	disableViewTableName := fmt.Sprintf("%s%s", tableName, appendStr)

	if !obj.UseView() {
		return disableViewTableName
	}

	if db.Statement != nil {
		disable, ok := db.InstanceGet(DisableViewsKey)
		if ok && disable.(bool) {
			return disableViewTableName
		}
	}

	view := fmt.Sprintf("%s_view_%s", tableName, obj.ViewVersion())
	return RewriteName(db, view) + appendStr
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

// CurrentViewName returns the current view name for a ViewModel using its
// ViewVersion(), after name overrides.
func CurrentViewName(db *gorm.DB, obj ViewModel) string {
	version, _ := strconv.Atoi(strings.TrimPrefix(obj.ViewVersion(), "v"))
	return RewriteName(db, DefaultViewName(db, obj, version))
}

func CustomViewName(db *gorm.DB, obj any, name string) string {
	tableName := plugins.TableName(db, obj)
	return fmt.Sprintf("%s_%s", tableName, name)
}
