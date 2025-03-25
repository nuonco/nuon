package pagination

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func NewOffsetPaginationPlugin() *offsetPaginationPlugin {
	return &offsetPaginationPlugin{}
}

type offsetPaginationPlugin struct {
	models []interface{}
}

func (m *offsetPaginationPlugin) Name() string {
	return "pagination-plugin"
}

func (m *offsetPaginationPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().Before("gorm:query").Register("enable_pagination_on_query", m.enablePagination)
	return nil
}

func (m *offsetPaginationPlugin) enablePagination(tx *gorm.DB) {
	enablePagination, ok := tx.Get(OffsetPaginationEnabledKey)
	if !(ok && enablePagination.(bool)) {
		return
	}

	ctxPagination := cctx.OffsetPaginationFromContext(tx.Statement.Context)

	if ctxPagination != nil {
		tx.Limit(ctxPagination.Limit + 1)
		tx.Offset(ctxPagination.Offset)
	}
}
