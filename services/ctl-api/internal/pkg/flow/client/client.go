package client

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Client provides methods for interacting with running flow workflows
// via Temporal update handlers. It is a direct Go client (not a Temporal
// activity) called from API handlers.
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

// findQueueSignalByOwner looks up the most recent queue signal for a given owner and signal type.
// The ownerType parameter is accepted for backwards compatibility but is not
// used in the query — the ownerID + signalType pair is sufficient to uniquely
// identify the signal regardless of which queue it was enqueued to.
func (c *Client) findQueueSignalByOwner(ctx context.Context, ownerID, ownerType string, signalType signal.SignalType) (*app.QueueSignal, error) {
	var qs app.QueueSignal
	res := c.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID: ownerID,
			Type:    signalType,
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("queue signal not found for owner %s type %s: %w", ownerID, signalType, res.Error)
	}
	return &qs, nil
}
