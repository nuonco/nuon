package pagination

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func NewPaginationPlugin() *paginationPlugin {
	return &paginationPlugin{}
}

type paginationPlugin struct {
	models []interface{}
}

func (m *paginationPlugin) Name() string {
	return "pagination-plugin"
}

func (m *paginationPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().Before("gorm:query").Register("enable_pagination_on_query", m.enablePagination)
	return nil
}

func (m *paginationPlugin) enablePagination(tx *gorm.DB) {
	enablePagination, ok := tx.Get(PaginationEnabledKey)
	if !(ok && enablePagination.(bool)) {
		return
	}

	ctxPagination := cctx.PaginationFromContext(tx.Statement.Context)

	if ctxPagination != nil {
		tx.Limit(ctxPagination.Limit + 1)
		tx.Offset(ctxPagination.Offset)
	}
}
