package client

import (
	"fmt"
	"sort"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func EnsureQueueSignal(ctx workflow.Context, ownerID, ownerType string, signalTypes ...signal.SignalType) error {
	sorted := make([]string, len(signalTypes))
	for i, st := range signalTypes {
		sorted[i] = string(st)
	}
	sort.Strings(sorted)
	cbID := fmt.Sprintf("ensure-%s-%s-%s", ownerType, ownerID, strings.Join(sorted, "+"))
	cbRef := callback.New(ctx, cbID)

	resp, err := AwaitEnsureSignal(ctx, &EnsureSignalRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		SignalTypes: signalTypes,
		Callback:    cbRef,
	})
	if err != nil {
		if dbgenerics.IsGormErrRecordNotFound(err) {
			return nil
		}
		return fmt.Errorf("ensure signal %v for %s/%s: %w", signalTypes, ownerType, ownerID, err)
	}

	if resp.AlreadyComplete {
		return nil
	}

	if _, err := callback.AwaitWithTimeout(ctx, cbRef, callback.FallbackAwaitTimeout); err != nil {
		return fmt.Errorf("ensure signal %v for %s/%s failed while awaiting: %w", signalTypes, ownerType, ownerID, err)
	}

	return nil
}
