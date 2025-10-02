package helpers

import (
	"context"
	"fmt"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) WriteJobPlan(ctx context.Context, jobID string, byts []byte, cp plantypes.CompositePlan) error {
	plan := app.RunnerJobPlan{
		RunnerJobID:   jobID,
		PlanJSON:      string(byts),
		CompositePlan: cp,
	}

	res := h.db.WithContext(ctx).Create(&plan)
	if res.Error != nil {
		return fmt.Errorf("unable to write job plan: %w", res.Error)
	}

	return nil
}
