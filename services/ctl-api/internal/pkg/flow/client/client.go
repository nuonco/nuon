package client

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

// findQueueSignalByOwner looks up the most recent queue signal for a given owner.
func (c *Client) findQueueSignalByOwner(ctx context.Context, ownerID, ownerType string) (*app.QueueSignal, error) {
	var qs app.QueueSignal
	res := c.db.WithContext(ctx).
		Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("queue signal not found for owner %s/%s: %w", ownerType, ownerID, res.Error)
	}
	return &qs, nil
}
