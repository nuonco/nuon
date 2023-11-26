package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/protos"
	"gorm.io/gorm"
)

type Activities struct {
	db     *gorm.DB
	protos *protos.Adapter
}

func New(prt *protos.Adapter, db *gorm.DB) (*Activities, error) {
	return &Activities{
		db:     db,
		protos: prt,
	}, nil
}
