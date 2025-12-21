package seed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (s *Seeder) EnsureCronEmitter(ctx context.Context, t *testing.T, queueID string, sig signal.Signal) *app.QueueEmitter {
	em, err := s.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        queueID,
		Name:           generics.GetFakeObj[string](),
		Description:    "test cron emitter",
		Mode:           app.QueueEmitterModeCron,
		CronSchedule:   "* * * * *", // Every minute
		SignalType:     sig.Type(),
		SignalTemplate: sig,
	})
	require.Nil(t, err)
	require.NotNil(t, em)

	return em
}

func (s *Seeder) EnsureScheduledEmitter(ctx context.Context, t *testing.T, queueID string, sig signal.Signal, scheduledAt time.Time) *app.QueueEmitter {
	em, err := s.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        queueID,
		Name:           generics.GetFakeObj[string](),
		Description:    "test scheduled emitter",
		Mode:           app.QueueEmitterModeScheduled,
		ScheduledAt:    &scheduledAt,
		SignalType:     sig.Type(),
		SignalTemplate: sig,
	})
	require.Nil(t, err)
	require.NotNil(t, em)

	return em
}
