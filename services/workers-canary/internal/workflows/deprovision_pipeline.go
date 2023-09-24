package workflows

import (
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) execDeprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) error {
	var runResp activities.RunTerraformResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.RunTerraform, &activities.RunTerraformRequest{
		RunType:  activities.RunTypeDestroy,
		CanaryID: req.CanaryId,
		OrgID:    req.OrgId,
	}, &runResp); err != nil {
		return fmt.Errorf("unable to run terraform: %w", err)
	}
	w.l.Info("run terraform", zap.Any("response", runResp))

	var orgResp activities.DeleteOrgResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.DeleteOrg, &activities.DeleteOrgRequest{
		CanaryID: req.CanaryId,
		OrgID:    req.OrgId,
	}, &orgResp); err != nil {
		return fmt.Errorf("unable to delete org: %w", err)
	}
	w.l.Info("deleted org", zap.Any("response", orgResp))

	return nil
}
