package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetLogStream(ctx context.Context, logStreamID string) (*models.AppLogStream, error) {
	resp, err := c.genClient.Operations.GetLogStream(&operations.GetLogStreamParams{
		LogStreamID: logStreamID,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) LogStreamReadLogs(ctx context.Context, logStreamId string, offset string, order string) ([]*models.AppOtelLogRecord, error) {
	params := &operations.LogStreamReadLogsParams{
		LogStreamID:    logStreamId,
		XNuonAPIOffset: &offset,
		Context:        ctx,
	}
	if order != "" {
		params.Order = &order
	}
	resp, err := c.genClient.Operations.LogStreamReadLogs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// LogStreamTailLogs hits the long-poll tail endpoint. `since` is a composite
// cursor (`<unix_nano>:<id>`, empty for "from oldest"); `wait` is an optional
// Go duration string, capped server-side at 30s when omitted.
//
// The endpoint is gated by the `log-tail-long-poll` org feature flag and
// returns 404 when the flag is off — callers should pick the tail vs legacy
// path up-front from `GetOrg().Features` rather than probing this endpoint.
func (c *client) LogStreamTailLogs(ctx context.Context, logStreamID string, since string, wait string) (*models.ServiceLogStreamTailLogsResponse, error) {
	params := &operations.LogStreamTailLogsParams{
		LogStreamID: logStreamID,
		Context:     ctx,
	}
	if since != "" {
		params.Since = &since
	}
	if wait != "" {
		params.Wait = &wait
	}
	resp, err := c.genClient.Operations.LogStreamTailLogs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) LogStreamReadLogsWithNextOffset(ctx context.Context, logStreamId string, offset string, order string) ([]*models.AppOtelLogRecord, string, error) {
	hr := newResponseHeaderReader(&operations.LogStreamReadLogsReader{})

	params := &operations.LogStreamReadLogsParams{
		LogStreamID:    logStreamId,
		XNuonAPIOffset: &offset,
		Context:        ctx,
	}
	if order != "" {
		params.Order = &order
	}
	resp, err := c.genClient.Operations.LogStreamReadLogs(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, "", err
	}

	nextOffset := hr.GetHeader("X-Nuon-API-Next")
	return resp.Payload, nextOffset, nil
}
