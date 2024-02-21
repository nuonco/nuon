package workflows

import (
	"fmt"
	"time"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const defaultDeleteWait = time.Hour * 4

func (w *wkflow) execDeprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) error {
	workflow.Sleep(ctx, defaultDeleteWait)

	var userResp activities.CreateUserResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateUser, &activities.CreateUserRequest{
		CanaryID: req.CanaryId,
	}, &userResp); err != nil {
		w.metricsWriter.Incr(ctx, "deprovision", 1, "status:error", "step:create_user")
		return fmt.Errorf("unable to create user: %w", err)
	}

	var getOrgResponse activities.GetOrgResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.GetOrg, &activities.GetOrgRequest{
		CanaryID: req.CanaryId,
	}, &getOrgResponse); err != nil {
		w.metricsWriter.Incr(ctx, "deprovision", 1, "status:error", "step:get_org")
		return fmt.Errorf("unable to get org: %w", err)
	}

	var runResp activities.RunTerraformResponse
	if err := w.defaultTerraformRunActivity(ctx, w.acts.RunTerraform, &activities.RunTerraformRequest{
		RunType:  activities.RunTypeDestroy,
		CanaryID: req.CanaryId,
		OrgID:    getOrgResponse.OrgID,
		APIToken: userResp.APIToken,
	}, &runResp, 3); err != nil {
		w.metricsWriter.Incr(ctx, "deprovision", 1, "status:error", "step:terraform_destroy")
		w.l.Info("error running terraform destroy", zap.Error(err))
	}
	w.l.Info("run terraform", zap.Any("response", runResp))

	// wait to delete the org to give us a chance to debug provision failures
	workflow.Sleep(ctx, defaultDeleteWait)

	var orgResp activities.DeleteOrgResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.DeleteOrg, &activities.DeleteOrgRequest{
		CanaryID: req.CanaryId,
		OrgID:    getOrgResponse.OrgID,
	}, &orgResp); err != nil {
		w.metricsWriter.Incr(ctx, "deprovision", 1, "status:error", "step:delete_org")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	w.l.Info("deleted org", zap.Any("response", orgResp))

	return nil
}
