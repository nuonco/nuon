package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

type PollSignalOptions struct {
	Timeout      *time.Duration
	PollInterval time.Duration
}

var ErrSignalTimeout = errors.New("timeout waiting for signal completion")

func parsePollSignalOptions(opts *PollSignalOptions) *PollSignalOptions {
	if opts == nil {
		opts = &PollSignalOptions{}
	}
	if opts.PollInterval == 0 {
		opts.PollInterval = 1 * time.Second
	}
	return opts
}

func (c *Client) PollSignal(ctx context.Context, queueSignalID string, opts *PollSignalOptions) (*handler.StatusResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	opts = parsePollSignalOptions(opts)

	pollCtx := ctx
	if opts.Timeout != nil {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, *opts.Timeout)
		defer cancel()
	}

	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			if opts.Timeout != nil && pollCtx.Err() == context.DeadlineExceeded {
				return nil, ErrSignalTimeout
			}
			return nil, pollCtx.Err()

		case <-ticker.C:
			resp, err := c.tClient.QueryWorkflowInNamespace(pollCtx, q.Workflow.Namespace, q.Workflow.ID, "", handler.StatusQueryName, &handler.StatusRequest{})
			if err != nil {
				return nil, errors.Wrap(err, "unable to query workflow status")
			}

			var status handler.StatusResponse
			if err := resp.Get(&status); err != nil {
				return nil, errors.Wrap(err, "unable to decode status response")
			}

			if status.Finished {
				return &status, nil
			}
		}
	}
}
