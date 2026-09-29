package client

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Client struct {
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
	L       *zap.Logger
}

func New(params Params) *Client {
	return &Client{
		db:      params.DB,
		tClient: params.TClient,
		l:       params.L,
	}
}

func (c *Client) findQueueSignalByOwner(ctx context.Context, ownerID, ownerType string, signalType signal.SignalType) (*app.QueueSignal, error) {
	var qs app.QueueSignal
	res := c.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID: ownerID,
			Type:    signalType,
		}).
		Order("created_at DESC").
		First(&qs)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) && signalType == executeflow.SignalType {
		res = c.db.WithContext(ctx).
			Where(app.QueueSignal{Type: signalType}).
			Where("signal->'data'->>'workflow_id' = ?", ownerID).
			Order("created_at DESC").
			First(&qs)
	}
	if res.Error != nil {
		return nil, fmt.Errorf("queue signal not found for owner %s type %s: %w", ownerID, signalType, res.Error)
	}
	return &qs, nil
}
