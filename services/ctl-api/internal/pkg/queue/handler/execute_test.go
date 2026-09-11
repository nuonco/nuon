package handler

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	queuecctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/cctx"
)

func TestSignalContextHydratesLogStreamOnceAcrossPhases(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))

	env.OnActivity(
		new(activities.Activities).HydrateLogStream,
		mock.Anything,
		mock.MatchedBy(func(req *activities.HydrateLogStreamRequest) bool {
			return req.LogStreamID == "log-stream-id" && req.OrgID == "org-id"
		}),
	).Return(&app.LogStream{
		ID:           "log-stream-id",
		OrgID:        "org-id",
		RunnerAPIURL: "https://runner.example.com",
		WriteToken:   "write-token",
	}, nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) (string, error) {
		h := &handler{
			queueSignal: &app.QueueSignal{
				SignalContext: queuecctx.SignalContext{
					OrgID:       "org-id",
					LogStreamID: "log-stream-id",
				},
			},
		}

		validateCtx, err := h.signalContext(ctx, true)
		if err != nil {
			return "", err
		}
		if _, err := cctx.GetLogStreamWorkflow(validateCtx); err != nil {
			return "", err
		}

		execCtx, err := h.signalContext(ctx, false)
		if err != nil {
			return "", err
		}
		stream, err := cctx.GetLogStreamWorkflow(execCtx)
		if err != nil {
			return "", err
		}
		return stream.RunnerAPIURL + "|" + stream.WriteToken, nil
	})

	require.NoError(t, env.GetWorkflowError())
	var credentials string
	require.NoError(t, env.GetWorkflowResult(&credentials))
	require.Equal(t, "https://runner.example.com|write-token", credentials)
	env.AssertExpectations(t)
}
