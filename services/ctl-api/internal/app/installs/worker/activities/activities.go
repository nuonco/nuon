package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/components"
	"gorm.io/gorm"
)

type Activities struct {
	db         *gorm.DB
	components *components.Adapter
}

func New(db *gorm.DB, comps *components.Adapter) (*Activities, error) {
	return &Activities{
		db:         db,
		components: comps,
	}, nil
}
