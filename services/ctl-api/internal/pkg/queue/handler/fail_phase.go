package handler

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// failPhase persists a failure of a signal phase (Validate or Execute) to
// the QueueSignal status, captures the StepErrorPayload-shaped metadata for
// the finishedHandler fast path, and returns a non-retryable
// ApplicationError that AwaitSignal will surface to the workflow caller.
//
// finishedAtKey is the Metadata key recording the phase completion
// timestamp ("validate_finished_at" or "execute_finished_at"). wrapErr is
// the phase-specific wrapper error (e.g. *signal.SignalErrValidate)
// passed as the cause on the returned ApplicationError.
//
// When err carries an stderr.ErrUser, its Description overrides the
// human-facing message and its code/fields/directive are persisted into
// the QueueSignal status metadata + attached as ApplicationError details
// so the workflow side can recover the typed error via
// extractUserError.
func (h *handler) failPhase(ctx workflow.Context, finishedAtKey string, wrapErr, err error) error {
	humanDesc := signal.HumanError(err)
	userDesc, userMeta, payload := extractStepDirective(err)
	if userDesc != "" {
		humanDesc = userDesc
	}

	meta := map[string]any{
		finishedAtKey: workflow.Now(ctx).UTC().Format(time.RFC3339),
	}
	for k, v := range userMeta {
		meta[k] = v
	}

	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID:     h.queueSignalID,
		Status:            app.StatusError,
		StatusDescription: humanDesc,
		Metadata:          meta,
	})
	// Pass the full merged meta (including the finishedAt timestamp) so
	// FinishedResponse.Metadata stays consistent with what we just
	// persisted to the DB; AwaitSignal's fast path relies on that
	// equivalence.
	h.setFinishedWithMeta(app.StatusError, humanDesc, meta)

	if payload.IsZero() {
		return temporal.NewNonRetryableApplicationError("signal failure", humanDesc, wrapErr)
	}
	return temporal.NewNonRetryableApplicationError("signal failure", humanDesc, wrapErr, payload)
}
