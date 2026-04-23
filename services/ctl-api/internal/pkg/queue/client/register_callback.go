package client

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/callback"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type RegisterCallbackRequest struct {
	QueueSignalID string                 `validate:"required"`
	Event         callback.Event         `validate:"required"`
	UpdateHandler signaldb.UpdateHandler `validate:"required"`
}

// RegisterCallback creates a callback for an existing queue signal.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) RegisterCallback(ctx context.Context, req *RegisterCallbackRequest) (*app.QueueSignalCallback, error) {
	cb := app.QueueSignalCallback{
		QueueSignalID: req.QueueSignalID,
		Event:         string(req.Event),
		UpdateHandler: req.UpdateHandler,
	}

	if res := c.db.WithContext(ctx).Create(&cb); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue signal callback")
	}

	return &cb, nil
}
