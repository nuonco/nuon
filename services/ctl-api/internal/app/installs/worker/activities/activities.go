package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/protos"
	"gorm.io/gorm"
)

type Activities struct {
	db         *gorm.DB
	components *protos.Adapter
}

func New(db *gorm.DB, prt *protos.Adapter) (*Activities, error) {
	return &Activities{
		db:         db,
		components: prt,
	}, nil
}
