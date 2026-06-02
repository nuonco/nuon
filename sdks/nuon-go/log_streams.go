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
