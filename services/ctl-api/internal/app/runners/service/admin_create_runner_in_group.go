package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnergroupssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runner_groups/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type createRunnerInGroupRequest struct {
	Platform app.AppRunnerType `json:"platform" validate:"required"`
}

type createRunnerInGroupResponse struct {
	Runner *app.Runner `json:"runner"`
	Token  string      `json:"token"`
}

// ensureRunnerServiceAccount finds or creates a service account for the runner
// and assigns it the runner org role. Failures are logged but not fatal.
func (s *service) ensureRunnerServiceAccount(ctx *gin.Context, runnerID, orgID string) {
	acct, err := s.acctClient.FindAccount(ctx, account.ServiceAccountEmail(runnerID))
	if err != nil {
		// No existing account — create one.
		acct, err = s.acctClient.CreateServiceAccount(ctx, runnerID)
		if err != nil {
			s.l.Warn("unable to create service account for runner",
				zap.String("runner_id", runnerID),
				zap.Error(err),
			)
			return
		}
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, orgID, acct.ID); err != nil {
		s.l.Warn("unable to add org role for runner",
			zap.String("runner_id", runnerID),
			zap.Error(err),
		)
	}
}

// @ID						AdminCreateRunnerInGroup
// @Summary				find or create a runner in a runner group
// @Description			Find an existing runner with matching platform or create a new one with service account and token
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Param					runner_group_id	path	string							true	"runner group ID"
// @Param					request			body	createRunnerInGroupRequest		true	"runner creation request"
// @Success				200	{object}	createRunnerInGroupResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/runner-groups/{runner_group_id}/runners [post]
func (s *service) AdminCreateRunnerInGroup(ctx *gin.Context) {
	groupID := ctx.Param("runner_group_id")

	var req createRunnerInGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid request: %w", err),
			Description: "invalid request body",
		})
		return
	}

	if req.Platform == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("platform is required"),
			Description: "platform field is required",
		})
		return
	}

	// Look up the runner group
	var group app.RunnerGroup
	res := s.db.WithContext(ctx).First(&group, "id = ? AND deleted_at = 0", groupID)
	if res.Error != nil {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("runner group %s not found: %w", groupID, res.Error),
			Description: "runner group not found",
		})
		return
	}

	// Find existing runner with matching platform in this group
	var existing app.Runner
	res = s.db.WithContext(ctx).
		Where("runner_group_id = ? AND platform = ? AND deleted_at = 0", groupID, req.Platform).
		First(&existing)
	if res.Error == nil {
		s.ensureRunnerServiceAccount(ctx, existing.ID, group.OrgID)

		// Ensure the runner's event loop is running (idempotent restart)
		s.evClient.Send(ctx, existing.ID, &signals.Signal{
			Type: signals.OperationRestart,
		})

		// Found existing runner, return it with a fresh token
		token, err := s.helpers.CreateToken(ctx, existing.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to create token for existing runner: %w", err))
			return
		}

		ctx.JSON(http.StatusOK, createRunnerInGroupResponse{
			Runner: &existing,
			Token:  token.Token,
		})
		return
	}

	// Create new runner — local runners are immediately active since they
	// don't go through the cloud provisioning workflow.
	status := app.RunnerStatusPending
	statusDesc := string(app.RunnerStatusPending)
	if req.Platform == app.AppRunnerTypeLocal {
		status = app.RunnerStatusActive
		statusDesc = "local runner is active"
	}

	runner := app.Runner{
		RunnerGroupID:     groupID,
		OrgID:             group.OrgID,
		Name:              string(req.Platform),
		DisplayName:       fmt.Sprintf("%s runner", req.Platform),
		Status:            status,
		StatusDescription: statusDesc,
		Platform:          req.Platform,
	}

	res = s.db.WithContext(ctx).Create(&runner)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create runner: %w", res.Error))
		return
	}

	s.ensureRunnerServiceAccount(ctx, runner.ID, group.OrgID)

	// Start the runner's event loop so it can receive job signals
	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	// Trigger leader election for the group asynchronously.
	s.evClient.Send(ctx, groupID, &runnergroupssignals.Signal{
		Type: runnergroupssignals.OperationElectLeader,
	})

	// Create token
	token, err := s.helpers.CreateToken(ctx, runner.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create token for runner: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, createRunnerInGroupResponse{
		Runner: &runner,
		Token:  token.Token,
	})
}
