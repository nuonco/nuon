package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CloseLogStreamRequest struct {
	LogStreamID string `validate:"required"`
}

// @temporal-gen activity
// @by-id LogStreamID
func (a *Activities) CloseLogStream(ctx context.Context, req CloseLogStreamRequest) error {
	ls := &app.LogStream{
		ID: req.LogStreamID,
	}
	res := a.db.WithContext(ctx).
		Model(&ls).
		Updates(map[string]interface{}{
			"open": false,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update")
	}

	return nil
}
