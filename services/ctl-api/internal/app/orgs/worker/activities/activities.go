package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"gorm.io/gorm"
)

type Activities struct {
	db *gorm.DB
}

func New(cfg *internal.Config,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db: db,
	}, nil
}
