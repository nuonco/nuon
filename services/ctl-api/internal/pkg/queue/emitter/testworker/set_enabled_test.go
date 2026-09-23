package testworker

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

const testOwnerType = "installs"

func (e *EmitterTestSuite) ensureOwnedQueue(ctx context.Context, ownerID string) *app.Queue {
	q, err := e.service.QueueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:   ownerID,
		OwnerType: testOwnerType,
		Namespace: defaultNamespace,
		MaxDepth:  100,
	})
	require.NoError(e.T(), err)

	e.T().Cleanup(func() {
		require.NoError(e.T(), e.service.QueueClient.Terminate(context.WithoutCancel(ctx), q.ID))
	})
	return q
}

// insertEmitter writes a row without going through CreateEmitter, so no
// Temporal workflow is started. Used for the cases that must be left alone —
// nothing should ever reach their (nonexistent) workflows.
func (e *EmitterTestSuite) insertEmitter(ctx context.Context, queueID string, mode app.QueueEmitterMode, deleted bool) *app.QueueEmitter {
	em := app.QueueEmitter{
		QueueID:      queueID,
		Name:         generics.GetFakeObj[string](),
		Mode:         mode,
		CronSchedule: "* * * * *",
		SignalType:   example.ExampleSignalType,
		Status:       app.NewCompositeStatus(ctx, app.StatusInProgress),
	}
	require.NoError(e.T(), e.service.DB.WithContext(ctx).Create(&em).Error)

	if deleted {
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Delete(&em).Error)
	}
	return &em
}

func (e *EmitterTestSuite) emitter(emitterID string) *app.QueueEmitter {
	var em app.QueueEmitter
	require.NoError(e.T(), e.service.DB.Unscoped().First(&em, "id = ?", emitterID).Error)
	return &em
}

func (e *EmitterTestSuite) TestSetCronEmittersEnabledForOwnerScope() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	ownerID := generics.GetFakeObj[string]()
	otherOwnerID := generics.GetFakeObj[string]()
	queue := e.ensureOwnedQueue(ctx, ownerID)
	otherQueue := e.ensureOwnedQueue(ctx, otherOwnerID)

	cron := e.service.Seed.EnsureCronEmitter(ctx, e.T(), queue.ID, &example.ExampleSignal{})
	fireOnce := e.insertEmitter(ctx, queue.ID, app.QueueEmitterModeFireOnce, false)
	deletedCron := e.insertEmitter(ctx, queue.ID, app.QueueEmitterModeCron, true)
	otherOwnerCron := e.insertEmitter(ctx, otherQueue.ID, app.QueueEmitterModeCron, false)

	disable := func() *emitterclient.SetCronEmittersEnabledResponse {
		resp, err := e.service.EmitterClient.SetCronEmittersEnabledForOwner(ctx, &emitterclient.SetCronEmittersEnabledRequest{
			OwnerID:   ownerID,
			OwnerType: testOwnerType,
			Reason:    "no healthy runner",
		})
		require.NoError(e.T(), err)
		return resp
	}

	resp := disable()
	require.Equal(e.T(), 1, resp.Changed)
	require.Equal(e.T(), []string{cron.ID}, resp.EmitterIDs)
	require.Zero(e.T(), resp.Errors, "the emitter's workflow should have accepted the stop update")

	disabled := e.emitter(cron.ID)
	require.Equal(e.T(), app.StatusDisabled, disabled.Status.Status)
	require.Equal(e.T(), "no healthy runner", disabled.Status.StatusHumanDescription)

	require.Equal(e.T(), app.StatusInProgress, e.emitter(fireOnce.ID).Status.Status, "fire-once emitters are not gated")
	require.Equal(e.T(), app.StatusInProgress, e.emitter(deletedCron.ID).Status.Status, "soft-deleted emitters must not be touched")
	require.Equal(e.T(), app.StatusInProgress, e.emitter(otherOwnerCron.ID).Status.Status, "another owner's emitters must not be touched")

	// A converged install must not keep re-issuing Temporal work every sweep.
	require.Equal(e.T(), 0, disable().Changed)
}

func (e *EmitterTestSuite) TestSetCronEmittersEnabledForOwnerReEnable() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	ownerID := generics.GetFakeObj[string]()
	queue := e.ensureOwnedQueue(ctx, ownerID)
	cron := e.service.Seed.EnsureCronEmitter(ctx, e.T(), queue.ID, &example.ExampleSignal{})

	_, err := e.service.EmitterClient.SetCronEmittersEnabledForOwner(ctx, &emitterclient.SetCronEmittersEnabledRequest{
		OwnerID:   ownerID,
		OwnerType: testOwnerType,
		Reason:    "no healthy runner",
	})
	require.NoError(e.T(), err)
	require.Equal(e.T(), app.StatusDisabled, e.emitter(cron.ID).Status.Status)

	resp, err := e.service.EmitterClient.SetCronEmittersEnabledForOwner(ctx, &emitterclient.SetCronEmittersEnabledRequest{
		OwnerID:   ownerID,
		OwnerType: testOwnerType,
		Enabled:   true,
		Reason:    "runner healthy",
	})
	require.NoError(e.T(), err)
	require.Equal(e.T(), 1, resp.Changed)

	reEnabled := e.emitter(cron.ID)
	require.Equal(e.T(), app.StatusInProgress, reEnabled.Status.Status)
	require.Empty(e.T(), reEnabled.Status.StatusHumanDescription, "the disable reason should not outlive the disable")
}
