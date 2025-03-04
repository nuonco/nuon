package scopes

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/pagination"
)

// EnablePagination is a scope that enables pagination for the query.
func WithPagination(db *gorm.DB) *gorm.DB {
	return db.Set(pagination.PaginationEnabledKey, true)
}
