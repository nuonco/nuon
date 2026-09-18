package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
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

// ResolveQueueByOwner picks an owner's queue without a name: its declared
// default, or its single queue for owner types that only ever have one. It is
// deliberately not an activity — callers name their queue. The one caller is
// EnqueueSignalToOwner, which still sees requests with no QueueName from
// workflows that were already in flight when this rolled out.
func (c *Client) ResolveQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error) {
	if _, ok := queuenames.Default(ownerType); ok {
		return c.GetDefaultQueueByOwner(ctx, ownerID, ownerType)
	}
	if queuenames.Sole(ownerType) {
		return c.GetOnlyQueueByOwner(ctx, ownerID, ownerType)
	}
	return nil, errors.Errorf("owner type %s has no default or sole queue", ownerType)
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetDefaultQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error) {
	name, ok := queuenames.Default(ownerType)
	if !ok {
		return nil, errors.Errorf("owner type %s has no default queue", ownerType)
	}

	var q app.Queue
	if res := c.defaultQueueByOwnerQuery(ctx, ownerID, ownerType, name).First(&q); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get default queue by owner")
	}

	return &q, nil
}

func (c *Client) defaultQueueByOwnerQuery(ctx context.Context, ownerID, ownerType, name string) *gorm.DB {
	return c.db.WithContext(ctx).
		Where(&app.Queue{OwnerID: ownerID, OwnerType: ownerType, Name: name}, "owner_id", "owner_type", "name")
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetOnlyQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error) {
	var queues []app.Queue
	if res := c.db.WithContext(ctx).
		Where(&app.Queue{OwnerID: ownerID, OwnerType: ownerType}).
		Limit(2).
		Find(&queues); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get only queue by owner")
	}
	if len(queues) == 0 {
		return nil, generics.TemporalGormError(gorm.ErrRecordNotFound, "unable to get only queue by owner")
	}
	if len(queues) != 1 {
		return nil, errors.Errorf("owner %s of type %s has %d queues, expected exactly one", ownerID, ownerType, len(queues))
	}
	return &queues[0], nil
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
		}, "owner_id", "owner_type", "name").
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
