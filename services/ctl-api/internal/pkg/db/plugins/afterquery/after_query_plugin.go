package afterquery

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type afterQuery interface {
	AfterQuery(db *gorm.DB) error
}

var _ gorm.Plugin = (*afterQueryPlugin)(nil)

func NewAfterQueryPlugin() *afterQueryPlugin {
	return &afterQueryPlugin{}
}

type afterQueryPlugin struct{}

func (d *afterQueryPlugin) Name() string {
	return "after-query"
}

func (d *afterQueryPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().After("gorm:query").Register("after_query", d.plugin)

	return nil
}

func (d *afterQueryPlugin) plugin(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	plugins.CallObjMethod(db, func(value interface{}, tx *gorm.DB) (called bool) {
		if i, ok := value.(afterQuery); ok {
			called = true
			db.AddError(i.AfterQuery(tx))
		}
		return called
	})
}
