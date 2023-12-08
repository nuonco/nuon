package workflows

import (
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) execProvision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*activities.TerraformRunOutputs, string, string, error) {
	var userResp activities.CreateUserResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateUser, &activities.CreateUserRequest{
		CanaryID: req.CanaryId,
	}, &userResp); err != nil {
		return nil, "", "", fmt.Errorf("unable to create user: %w", err)
	}

	var orgResp activities.CreateOrgResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateOrg, &activities.CreateOrgRequest{
		CanaryID:    req.CanaryId,
		SandboxMode: req.SandboxMode,
		APIToken:    userResp.APIToken,
	}, &orgResp); err != nil {
		return nil, "", "", fmt.Errorf("unable to create org: %w", err)
	}
	w.l.Info("create org", zap.Any("response", orgResp))

	var vcsResp activities.CreateVCSConnectionResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateVCSConnection, &activities.CreateVCSConnectionRequest{
		CanaryID:        req.CanaryId,
		APIToken:        userResp.APIToken,
		GithubInstallID: userResp.GithubInstallID,
		OrgID:           orgResp.OrgID,
	}, &vcsResp); err != nil {
		return nil, orgResp.OrgID, userResp.APIToken, fmt.Errorf("unable to create vcs connection: %w", err)
	}
	w.l.Info("create vcs connection", zap.Any("response", vcsResp))

	var runResp activities.RunTerraformResponse
	if err := w.defaultTerraformRunActivity(ctx, w.acts.RunTerraform, &activities.RunTerraformRequest{
		RunType:  activities.RunTypeApply,
		APIToken: userResp.APIToken,
		CanaryID: req.CanaryId,
		OrgID:    orgResp.OrgID,
	}, &runResp, 1); err != nil {
		return nil, orgResp.OrgID, userResp.APIToken, fmt.Errorf("unable to run terraform: %w", err)
	}
	w.l.Info("run terraform", zap.Any("response", runResp))

	return runResp.Outputs, orgResp.OrgID, userResp.APIToken, nil
}
