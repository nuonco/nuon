package activities

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
)

type Activities struct {
	*activities.Activities

	db       *gorm.DB
	wpClient multi.Client
}

func New(cfg *internal.Config,
	wpClient multi.Client,
	db *gorm.DB,
	acts *activities.Activities,
) (*Activities, error) {
	return &Activities{
		Activities: acts,
		db:         db,
		wpClient:   wpClient,
	}, nil
}
