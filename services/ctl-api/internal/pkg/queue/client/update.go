package client

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateQueueRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`
	Namespace string `validate:"required"`

	Name     string
	Metadata pgtype.Hstore

	MaxInFlight int
	MaxDepth    int
}

// Update finds an existing queue by owner ID, owner type, and name. If found,
// it updates the queue fields. If not found, it creates a new queue via Create.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Update(ctx context.Context, req *UpdateQueueRequest) (*app.Queue, error) {
	var existing app.Queue
	res := c.db.WithContext(ctx).
		Where("owner_id = ? AND owner_type = ? AND name = ?", req.OwnerID, req.OwnerType, req.Name).
		First(&existing)
	if res.Error == nil {
		existing.Metadata = req.Metadata
		existing.MaxInFlight = req.MaxInFlight
		existing.MaxDepth = req.MaxDepth

		if res := c.db.WithContext(ctx).Save(&existing); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to update queue")
		}

		c.l.Debug("queue updated",
			zap.String("id", existing.ID),
			zap.String("name", req.Name),
		)
		return &existing, nil
	}

	// Not found — create via existing method
	return c.Create(ctx, &CreateQueueRequest{
		OwnerID:     req.OwnerID,
		OwnerType:   req.OwnerType,
		Namespace:   req.Namespace,
		Name:        req.Name,
		Metadata:    req.Metadata,
		MaxInFlight: req.MaxInFlight,
		MaxDepth:    req.MaxDepth,
	})
}
