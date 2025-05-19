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
	compCount, err := activities.AwaitGetUnknownComponentCount(ctx, activities.GetcomponentRecordsCount{})
	if err != nil {
		return err
	}

	var batchSize int64 = 10
	batchCount := compCount / batchSize
	if compCount%batchSize != 0 {
		batchCount++
	}

	for i := int64(0); i < batchCount; i++ {
		impactedComps, err := activities.AwaitGetUnknownComponents(ctx, activities.GetUnknownComponents{
			Limit: int(batchSize),
		})
		if err != nil {
			return err
		}

		impactedIDs := make([]string, 0, len(impactedComps))
		for _, comp := range impactedComps {
			impactedIDs = append(impactedIDs, comp.ID)
		}

		if len(impactedComps) == 0 {
			break
		}

		fmt.Println("Processing batch: ", i+1, " with components: ", len(impactedComps))

		compsWithType, err := activities.AwaitGetComponentsWithType(ctx, activities.GetComponentsWithType{
			IDs: impactedIDs,
		})
		if err != nil {
			return err
		}
		fmt.Println("Found components with type: ", len(compsWithType))

		compType := make(map[string]app.ComponentType)
		for _, comp := range compsWithType {
			compType[comp.ID] = comp.Type
		}

		for _, comp := range impactedComps {
			comp.Type = compType[comp.ID]
			err := activities.AwaitUpdateComponentType(ctx, activities.UpdateComponentTypeRequest{
				ComponentID: comp.ID,
				Type:        comp.Type,
			})
			if err != nil {
				return err
			}
		}

	}

	return nil
}
