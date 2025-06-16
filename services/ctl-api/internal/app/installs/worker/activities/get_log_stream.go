package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLogStreamRequest struct {
	LogStreamID string `validate:"required"`
}

// @temporal-gen activity
// @by-id LogStreamID
func (a *Activities) GetLogStream(ctx context.Context, req GetLogStreamRequest) (*app.LogStream, error) {
	return a.getLogStream(ctx, req.LogStreamID)
}

func (a *Activities) getLogStream(ctx context.Context, logStreamID string) (*app.LogStream, error) {
	installLogStream := app.LogStream{}
	res := a.db.WithContext(ctx).First(&installLogStream, "id = ?", logStreamID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installLogStream, nil
}
