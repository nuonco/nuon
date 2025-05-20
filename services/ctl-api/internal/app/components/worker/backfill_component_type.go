package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
// This workflow is written as a temporary/one-time operation & is being used to backfill data.
func (w *Workflows) BackfillComponentType(ctx workflow.Context, sreq signals.RequestSignal) error {
	impactedCompIDs, err := activities.AwaitGetUnknownComponentIDs(ctx, activities.GetUnknownComponents{})
	if err != nil {
		return err
	}

	if len(impactedCompIDs) == 0 {
		return nil
	}

	batchSize := 1
	batchCount := len(impactedCompIDs) / batchSize
	if len(impactedCompIDs)%batchSize != 0 {
		batchCount++
	}

	fmt.Println("Found components with unknown type: ", len(impactedCompIDs))

	for i := 0; i < batchCount; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(impactedCompIDs) {
			end = len(impactedCompIDs)
		}
		batchIDs := impactedCompIDs[start:end]

		fmt.Printf("Batch %d/%d\n", i+1, batchCount)

		compsWithType, err := activities.AwaitGetComponentsWithType(ctx, activities.GetComponentsWithType{
			IDs: batchIDs,
		})
		if err != nil {
			return err
		}
		fmt.Println("Found components with type: ", len(compsWithType))

		if len(compsWithType) == 0 {
			continue
		}

		compType := make(map[string]app.ComponentType)
		for _, comp := range compsWithType {
			compType[comp.ID] = comp.Type
		}

		for _, compID := range batchIDs {
			if compType[compID] == "" {
				continue
			}
			err := activities.AwaitUpdateComponentType(ctx, activities.UpdateComponentTypeRequest{
				ComponentID: compID,
				Type:        compType[compID],
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
