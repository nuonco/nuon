package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/components"
	"gorm.io/gorm"
)

type Activities struct {
	db         *gorm.DB
	components *components.Adapter
}

func New(cfg *internal.Config,
	components *components.Adapter,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db: db,
	}, nil
}
