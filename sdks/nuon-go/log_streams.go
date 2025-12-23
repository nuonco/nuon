package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
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

func (c *client) LogStreamReadLogs(ctx context.Context, logStreamId string, offset string) ([]*models.AppOtelLogRecord, error) {
	resp, err := c.genClient.Operations.LogStreamReadLogs(&operations.LogStreamReadLogsParams{
		LogStreamID: logStreamId, XNuonAPIOffset: &offset},
		c.getOrgIDAuthInfo(),
	)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}
