package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCreateEmitterRejectsMissingSignalTemplate(t *testing.T) {
	c := &Client{}
	_, err := c.CreateEmitter(context.Background(), &CreateEmitterRequest{
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: "* * * * *",
	})
	require.Error(t, err)

	var appErr *temporal.ApplicationError
	require.True(t, errors.As(err, &appErr))
	require.True(t, appErr.NonRetryable())
	require.Equal(t, "EMITTER_CONFIG_ERROR", appErr.Type())
}
