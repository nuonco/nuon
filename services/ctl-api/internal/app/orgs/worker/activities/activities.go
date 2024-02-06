package activities

import (
	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"gorm.io/gorm"
)

type Activities struct {
	db       *gorm.DB
	wpClient multi.Client
}

func New(cfg *internal.Config,
	wpClient multi.Client,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db:       db,
		wpClient: wpClient,
	}, nil
}
