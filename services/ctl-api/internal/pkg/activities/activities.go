package activities

import (
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

type Activities struct {
	v      *validator.Validate
	db     *gorm.DB
	notifs *notifications.Notifications
}

func New(cfg *internal.Config,
	v *validator.Validate,
	notifs *notifications.Notifications,
	db *gorm.DB,
) (*Activities, error) {
	return &Activities{
		v:      v,
		db:     db,
		notifs: notifs,
	}, nil
}
