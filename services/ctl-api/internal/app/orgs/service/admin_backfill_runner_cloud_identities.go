package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type BackfillRunnerCloudIdentitiesResponse struct {
	Installs    int      `json:"installs"`
	Bound       int      `json:"bound"`
	Skipped     []string `json:"skipped"`
	WithoutRole []string `json:"without_role"`
}

// @ID						AdminBackfillOrgRunnerCloudIdentities
// @Summary				backfill runner cloud identities for an org
// @Description			Creates the runner cloud identity binding for every AWS install runner in the org that does not have one. The binding is the account, region and runner role that POST /v1/runner-auth/aws/sts checks a runner against. The account and region come from the install; the runner role is the runner_iam_role_arn in the install stack's outputs and is only bound when it is an IAM role in the install's account. An install with no target account is skipped and listed in the response; run the phone home backfill first to give it one. An install whose stack has not reported a usable runner role is listed in without_role. Idempotent.
// @Param					org_id	path	string	true	"org ID"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	BackfillRunnerCloudIdentitiesResponse
// @Router					/v1/orgs/{org_id}/admin-backfill-runner-cloud-identities [POST]
func (s *service) AdminBackfillOrgRunnerCloudIdentities(ctx *gin.Context) {
	org, err := s.getOrg(ctx, ctx.Param("org_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var installs []app.Install
	if err := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("InstallStack.InstallStackOutputs").
		Where(app.Install{OrgID: org.ID}).
		Find(&installs).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get installs: %w", err))
		return
	}

	resp := BackfillRunnerCloudIdentitiesResponse{Skipped: []string{}, WithoutRole: []string{}}
	installOwnerType := plugins.TableName(s.db, app.Install{})
	for i := range installs {
		install := &installs[i]
		if install.AWSAccount == nil {
			continue
		}
		resp.Installs++

		var runners []app.Runner
		if err := s.db.WithContext(ctx).
			Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id AND runner_groups.deleted_at = 0").
			Where("runner_groups.owner_id = ? AND runner_groups.owner_type = ?", install.ID, installOwnerType).
			Find(&runners).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to get runners for install %s: %w", install.ID, err))
			return
		}

		for _, runner := range runners {
			identity, err := s.runnersHelpers.EnsureInstallRunnerCloudIdentity(ctx, install, runner.ID)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to ensure runner cloud identity for install %s: %w", install.ID, err))
				return
			}
			if identity == nil {
				resp.Skipped = append(resp.Skipped, install.ID)
				continue
			}
			resp.Bound++

			if install.InstallStack == nil || install.InstallStack.InstallStackOutputs.AWSStackOutputs == nil {
				resp.WithoutRole = append(resp.WithoutRole, install.ID)
				continue
			}
			update, err := s.runnersHelpers.SetRunnerCloudIdentityRole(ctx, runner.ID, install.InstallStack.InstallStackOutputs.AWSStackOutputs.RunnerIAMRoleARN)
			if err != nil {
				ctx.Error(fmt.Errorf("unable to set runner role for install %s: %w", install.ID, err))
				return
			}
			if !update.Accepted {
				resp.WithoutRole = append(resp.WithoutRole, install.ID)
			}
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
