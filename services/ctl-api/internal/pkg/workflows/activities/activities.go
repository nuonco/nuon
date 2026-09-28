package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	Cfg         *internal.Config
	V           *validator.Validate
	Notifs      *notifications.Notifications
	DB          *gorm.DB `name:"psql"`
	QueueClient *client.Client
}

type Activities struct {
	cfg         *internal.Config
	v           *validator.Validate
	db          *gorm.DB
	notifs      *notifications.Notifications
	queueClient *client.Client
}

func New(params Params) (*Activities, error) {
	return &Activities{
		cfg:         params.Cfg,
		v:           params.V,
		db:          params.DB,
		notifs:      params.Notifs,
		queueClient: params.QueueClient,
	}, nil
}
