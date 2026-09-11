package queuecctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	qcctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/cctx"
)

func TestApplyDoesNotInstallPartialLogStream(t *testing.T) {
	ctx := Apply(context.Background(), qcctx.SignalContext{LogStreamID: "log-stream-id"})

	_, err := cctx.GetLogStreamContext(ctx)
	require.Error(t, err)
}

func TestApplyWorkflowDoesNotInstallPartialLogStream(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(func(ctx workflow.Context) (bool, error) {
		ctx = ApplyWorkflow(ctx, qcctx.SignalContext{LogStreamID: "log-stream-id"})
		_, err := cctx.GetLogStreamWorkflow(ctx)
		return err != nil, nil
	})

	require.NoError(t, env.GetWorkflowError())
	var missing bool
	require.NoError(t, env.GetWorkflowResult(&missing))
	require.True(t, missing)
}

func TestFromContextCapturesLogStreamIDWithoutCredentials(t *testing.T) {
	ctx := cctx.SetLogStreamContext(context.Background(), &app.LogStream{
		ID:           "log-stream-id",
		RunnerAPIURL: "https://runner.example.com",
		WriteToken:   "secret-token",
	})

	signalCtx := FromContext(ctx)

	require.Equal(t, "log-stream-id", signalCtx.LogStreamID)
}
