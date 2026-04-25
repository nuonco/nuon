package client

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// Status queries the state-manager workflow for its current metadata.
func (c *Client) Status(ctx context.Context, installID string) (*statemanager.StatusResponse, error) {
	wfID := workflowID(installID)
	val, err := c.tClient.QueryWorkflowInNamespace(ctx, workflowNamespace, wfID, "", statemanager.StatusQueryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query state-manager status")
	}

	var resp statemanager.StatusResponse
	if err := val.Get(&resp); err != nil {
		// Try json decode as fallback.
		var raw json.RawMessage
		if err2 := val.Get(&raw); err2 != nil {
			return nil, errors.Wrap(err, "unable to decode status response")
		}
		if err2 := json.Unmarshal(raw, &resp); err2 != nil {
			return nil, errors.Wrap(err, "unable to decode status response")
		}
	}
	return &resp, nil
}
