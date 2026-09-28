package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/reverify"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const ReverifySchedule = "0 * * * *"

func (h *Helpers) EnsureConnectionQueue(ctx context.Context, connection *app.CloudConnection) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID: connection.ID, OwnerType: queuenames.OwnerCloudConnections, Namespace: "orgs",
		Name: "cloud-connection-" + connection.ID, MaxInFlight: 1, MaxDepth: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("ensure cloud connection queue: %w", err)
	}
	if err := h.db.WithContext(ctx).Model(q).Select("idle_timeout").Updates(app.Queue{IdleTimeout: int64(5 * time.Second)}).Error; err != nil {
		return nil, err
	}
	emitterName := "cloud-connection-" + connection.ID + "-reverify"
	var existing app.QueueEmitter
	err = h.db.WithContext(ctx).Where(app.QueueEmitter{QueueID: q.ID, Name: emitterName}).First(&existing).Error
	if err == nil {
		err = h.db.WithContext(ctx).Model(&existing).Select("cron_schedule", "jitter_window", "signal_expires_in").Updates(app.QueueEmitter{
			CronSchedule: cronutil.ApplyCronJitter(existing.ID, ReverifySchedule, cronutil.MaxJitterWindow),
			JitterWindow: cronutil.MaxJitterWindow, SignalExpiresIn: 5 * time.Minute,
		}).Error
		return q, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	_, err = h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID: q.ID, Name: emitterName, Description: "Periodic cloud connection verification",
		Mode: app.QueueEmitterModeCron, CronSchedule: ReverifySchedule, JitterWindow: cronutil.MaxJitterWindow,
		SignalExpiresIn: 5 * time.Minute, SignalType: reverify.SignalType,
		SignalTemplate: &reverify.Signal{CloudConnectionID: connection.ID},
	})
	return q, err
}

func (h *Helpers) EnqueueVerification(ctx context.Context, connection *app.CloudConnection) error {
	q, err := h.EnsureConnectionQueue(ctx, connection)
	if err != nil {
		return err
	}
	_, err = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID, OwnerID: connection.ID, OwnerType: queuenames.OwnerCloudConnections,
		Signal: &reverify.Signal{CloudConnectionID: connection.ID, OnDemand: true},
	})
	return err
}

func (h *Helpers) TerminateConnectionQueue(ctx context.Context, connectionID string) error {
	var queues []app.Queue
	if err := h.db.WithContext(ctx).Where(app.Queue{OwnerID: connectionID, OwnerType: queuenames.OwnerCloudConnections}).Find(&queues).Error; err != nil {
		return err
	}
	for _, q := range queues {
		if err := h.queueClient.TerminateStrict(ctx, q.ID); err != nil {
			return err
		}
	}
	return nil
}
