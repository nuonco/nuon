package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type ForceRunnersSandboxModeRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) ForceRunnersSandboxMode(ctx context.Context, req ForceSandboxModeRequest) error {
	if err := a.forceRunnersSandboxMode(ctx, req.OrgID); err != nil {
		return errors.Wrap(err, "unable to force sandbox mode")
	}

	return nil
}

func (a *Activities) forceRunnersSandboxMode(ctx context.Context, orgID string) error {
	org := &app.RunnerGroupSettings{}
	res := a.db.WithContext(ctx).
		Model(org).
		Where(app.RunnerGroupSettings{
			OrgID: orgID,
		}).
		Updates(map[string]any{
			"sandbox_mode":                 true,
			"org_aws_iam_role_arn":         "",
			"org_k8s_service_account_name": "",
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to force sandbox mode for runner group settings")
	}

	return nil
}
