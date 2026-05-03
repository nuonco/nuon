package state

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
	L           *zap.Logger
}

type Activities struct {
	db          *gorm.DB
	queueClient *queueclient.Client
	l           *zap.Logger
}

func New(params Params) *Activities {
	return &Activities{
		db:          params.DB,
		queueClient: params.QueueClient,
		l:           params.L,
	}
}

func (a *Activities) getStateManagerQueueID(ctx context.Context, installID string) (string, error) {
	var q app.Queue
	if res := a.db.WithContext(ctx).
		Where(app.Queue{OwnerID: installID, Name: installshelpers.InstallStateManagerQueueName}).
		First(&q); res.Error != nil {
		return "", fmt.Errorf("unable to get state-manager queue for install %s: %w", installID, res.Error)
	}
	return q.ID, nil
}
