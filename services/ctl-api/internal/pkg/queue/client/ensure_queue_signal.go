package client

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnsureQueueSignal is a workflow-level helper that checks whether the latest
// signal of a given type for an owner has completed. If it has, it returns
// immediately. If it is still in flight, it registers a callback and blocks
// until the signal completes (or times out).
//
// This enables dependent signals to wait for prerequisites without polling,
// e.g. a component-build signal waiting for org-provision to finish.
func EnsureQueueSignal(ctx workflow.Context, ownerID, ownerType string, signalType signal.SignalType) error {
	cbID := fmt.Sprintf("ensure-%s-%s-%s", ownerType, ownerID, signalType)
	cbRef := callback.New(ctx, cbID)

	resp, err := AwaitEnsureSignal(ctx, &EnsureSignalRequest{
		OwnerID:    ownerID,
		OwnerType:  ownerType,
		SignalType: signalType,
		Callback:   cbRef,
	})
	if err != nil {
		if dbgenerics.IsGormErrRecordNotFound(err) {
			return nil
		}
		return fmt.Errorf("ensure signal %s for %s/%s: %w", signalType, ownerType, ownerID, err)
	}

	if resp.AlreadyComplete {
		return nil
	}

	// Block until the handler sends a completion callback.
	if _, err := callback.Await(ctx, cbRef); err != nil {
		return fmt.Errorf("ensure signal %s for %s/%s failed while awaiting: %w", signalType, ownerType, ownerID, err)
	}

	return nil
}
