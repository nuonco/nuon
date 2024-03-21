package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/metrics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) getInstallCount(sandboxMode bool) int {
	if sandboxMode {
		return w.cfg.SandboxModeInstallCount
	}

	return w.cfg.DefaultInstallCount
}

func (w *wkflow) execProvision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*activities.TerraformRunOutputs, string, string, error) {
	var userResp activities.CreateUserResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateUser, &activities.CreateUserRequest{
		CanaryID: req.CanaryId,
	}, &userResp); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:create_user", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
		return nil, "", "", fmt.Errorf("unable to create user: %w", err)
	}

	var orgResp activities.CreateOrgResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateOrg, &activities.CreateOrgRequest{
		CanaryID:    req.CanaryId,
		SandboxMode: req.SandboxMode,
		APIToken:    userResp.APIToken,
	}, &orgResp); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:create_org", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
		return nil, "", "", fmt.Errorf("unable to create org: %w", err)
	}
	w.l.Info("create org", zap.Any("response", orgResp))

	var addSupportUsersResp activities.AddSupportUsersResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.AddSupportUsers, &activities.AddSupportUsersRequest{
		OrgID: orgResp.OrgID,
	}, &addSupportUsersResp); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:add_support_users", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
		return nil, orgResp.OrgID, userResp.APIToken, fmt.Errorf("unable to add support users: %w", err)
	}
	w.l.Info("create support users", zap.Any("response", addSupportUsersResp))

	var vcsResp activities.CreateVCSConnectionResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateVCSConnection, &activities.CreateVCSConnectionRequest{
		CanaryID:        req.CanaryId,
		APIToken:        userResp.APIToken,
		GithubInstallID: userResp.GithubInstallID,
		OrgID:           orgResp.OrgID,
	}, &vcsResp); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:create_vcs_connection", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
		return nil, orgResp.OrgID, userResp.APIToken, fmt.Errorf("unable to create vcs connection: %w", err)
	}
	w.l.Info("create vcs connection", zap.Any("response", vcsResp))

	var runResp activities.RunTerraformResponse
	if err := w.defaultTerraformRunActivity(ctx, w.acts.RunTerraform, &activities.RunTerraformRequest{
		RunType:      activities.RunTypeApply,
		APIToken:     userResp.APIToken,
		CanaryID:     req.CanaryId,
		OrgID:        orgResp.OrgID,
		InstallCount: w.getInstallCount(req.SandboxMode),
	}, &runResp, 2); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:run_terraform", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
		return nil, orgResp.OrgID, userResp.APIToken, fmt.Errorf("unable to run terraform: %w", err)
	}
	w.l.Info("run terraform", zap.Any("response", runResp))

	return runResp.Outputs, orgResp.OrgID, userResp.APIToken, nil
}
