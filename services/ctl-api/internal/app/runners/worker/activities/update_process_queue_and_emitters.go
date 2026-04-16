package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateProcessQueueAndEmittersRequest struct {
	RunnerID           string `validate:"required"`
	ProcessID          string `validate:"required"`
	ContainerMaxUptime int    `validate:"required"`
	VMMaxUptime        int    `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) UpdateProcessQueueAndEmittersRequest(ctx context.Context, req UpdateProcessQueueAndEmittersRequest) error {
	var process app.RunnerProcess
	if res := a.db.WithContext(ctx).Where("id = ?", req.ProcessID).First(&process); res.Error != nil {
		return errors.Wrap(res.Error, fmt.Sprintf("unable to find runner process %s", req.ProcessID))
	}

	_, err := a.helpers.UpdateProcessQueues(ctx, req.RunnerID, &process)
	return err
}
