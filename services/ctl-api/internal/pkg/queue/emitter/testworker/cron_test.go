package testworker

import (
	"context"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	temporalclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EmitterTestSuite) TestCronTickEmitsSignal() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	queue := e.service.Seed.EnsureQueue(ctx, e.T())

	sig := &example.ExampleSignal{Arg1: "first", Arg2: "second"}
	em := e.service.Seed.EnsureCronEmitter(ctx, e.T(), queue.ID, sig)
	e.T().Cleanup(func() {
		require.NoError(e.T(), e.service.QueueClient.Terminate(context.WithoutCancel(ctx), queue.ID))
	})

	run, err := e.service.TemporalClient.ExecuteWorkflowInNamespace(ctx, defaultNamespace, temporalclient.StartWorkflowOptions{
		ID:                    "test-cron-ticker-54", // Hashes to zero seconds of deterministic cron jitter.
		TaskQueue:             queue.Workflow.TaskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}, "CronTicker", emitter.CronTickerWorkflowRequest{
		QueueID:   queue.ID,
		EmitterID: em.ID,
	})
	require.NoError(e.T(), err)
	require.NoError(e.T(), run.Get(ctx, nil))

	var emitted app.QueueSignal
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Where(app.QueueSignal{
		QueueID:   queue.ID,
		EmitterID: &em.ID,
	}).First(&emitted).Error)
	require.Equal(e.T(), example.ExampleSignalType, emitted.Type)
}
