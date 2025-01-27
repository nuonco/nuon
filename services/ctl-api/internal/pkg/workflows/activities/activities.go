package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

type Params struct {
	fx.In

	Cfg    *internal.Config
	V      *validator.Validate
	Notifs *notifications.Notifications
	DB     *gorm.DB `name:"psql"`
}

type Activities struct {
	v      *validator.Validate
	db     *gorm.DB
	notifs *notifications.Notifications
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:      params.V,
		db:     params.DB,
		notifs: params.Notifs,
	}, nil
}
