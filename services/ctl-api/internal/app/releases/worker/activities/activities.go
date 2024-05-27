package activities

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Activities struct {
	db       *gorm.DB
	evClient eventloop.Client
}

func New(cfg *internal.Config,
	evClient eventloop.Client,
	db *gorm.DB,
) (*Activities, error) {
	return &Activities{
		db:       db,
		evClient: evClient,
	}, nil
}
