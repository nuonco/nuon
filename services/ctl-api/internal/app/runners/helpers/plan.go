package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

func (h *Helpers) WriteJobPlan(ctx context.Context, jobID string, byts []byte, cp plantypes.CompositePlan) error {
	plan := app.RunnerJobPlan{
		RunnerJobID:   jobID,
		PlanJSON:      string(byts),
		CompositePlan: cp,
	}

	cpJSON, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("unable to marshal composite plan for blob: %w", err)
	}
	plan.CompositePlanBlob = &blobstore.Blob{}
	plan.CompositePlanBlob.Set(string(cpJSON))

	res := h.db.WithContext(ctx).Create(&plan)
	if res.Error != nil {
		return fmt.Errorf("unable to write job plan: %w", res.Error)
	}

	return nil
}
