package activities

import (
	"context"
	"fmt"
	"time"

	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

type NotifyEventRunbookWaiterRequest struct{ WaiterID, OrgID, QueueSignalID string }

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) NotifyEventRunbookWaiter(ctx context.Context, req NotifyEventRunbookWaiterRequest) error {
	var waiter app.EventRunbookWaiter
	if err := a.db.WithContext(ctx).Where(app.EventRunbookWaiter{ID: req.WaiterID, OrgID: req.OrgID}).First(&waiter).Error; err != nil {
		return err
	}
	if waiter.NotifiedAt != nil {
		return nil
	}
	if waiter.QueueSignalID != req.QueueSignalID {
		return temporal.NewNonRetryableApplicationError("queue signal does not belong to waiter", "waiter_queue_signal_mismatch", nil)
	}
	var qs app.QueueSignal
	if err := a.db.WithContext(ctx).Where(app.QueueSignal{ID: req.QueueSignalID}).First(&qs).Error; err != nil {
		return err
	}
	h, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{UpdateName: "event-received", WaitForStage: tclient.WorkflowUpdateStageCompleted})
	if err != nil {
		return fmt.Errorf("send event-received update: %w", err)
	}
	var result struct{}
	if err := h.Get(ctx, &result); err != nil {
		return err
	}
	now := time.Now().UTC()
	return a.db.WithContext(ctx).Model(&app.EventRunbookWaiter{}).Where(app.EventRunbookWaiter{ID: waiter.ID, Status: app.EventRunbookWaiterStatusMatched}).Update("notified_at", now).Error
}
