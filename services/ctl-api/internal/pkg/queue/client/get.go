package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetQueue(ctx context.Context, id string) (*app.Queue, error) {
	q, err := c.getQueue(ctx, id)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get queue")
	}

	return q, nil
}

func (c *Client) getQueue(ctx context.Context, id string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	return &q, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   ownerID,
			OwnerType: ownerType,
		}).
		First(&q); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get queue by owner")
	}

	return &q, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetQueueByOwnerAndName(ctx context.Context, ownerID, ownerType, name string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   ownerID,
			OwnerType: ownerType,
			Name:      name,
		}).
		First(&q); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get queue by owner and name")
	}

	return &q, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetQueueStatus(ctx context.Context, queueID string) (*queue.StatusResponse, error) {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.StatusHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.StatusRequest{},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call status handler")
	}

	var resp queue.StatusResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}

	return &resp, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ListQueues(ctx context.Context, orgID, ownerID, ownerType string, limit, offset int) ([]app.Queue, error) {
	query := c.db.WithContext(ctx).Where("org_id = ?", orgID)

	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if ownerType != "" {
		query = query.Where("owner_type = ?", ownerType)
	}

	var queues []app.Queue
	if res := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&queues); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list queues")
	}

	return queues, nil
}
