package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnergroupssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runner_groups/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// setRunnerTainted fetches a runner by ID, optionally scoped to an org, and sets its tainted field.
// When orgID is non-empty, the lookup includes an org_id filter and returns stderr.ErrNotFound on miss.
// When orgID is empty (admin path), the lookup omits the org filter and returns a plain error on miss.
func (s *service) setRunnerTainted(ctx *gin.Context, runnerID string, tainted bool, orgID string) (*app.Runner, error) {
	var runner app.Runner

	query := s.db.WithContext(ctx)
	if orgID != "" {
		query = query.Where("id = ? AND org_id = ?", runnerID, orgID)
	} else {
		query = query.Where("id = ?", runnerID)
	}

	if res := query.First(&runner); res.Error != nil {
		if orgID != "" {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("runner %s not found: %w", runnerID, res.Error),
				Description: "runner not found",
			}
		}
		return nil, fmt.Errorf("runner %s not found: %w", runnerID, res.Error)
	}

	if res := s.db.WithContext(ctx).Model(&runner).Update("tainted", tainted); res.Error != nil {
		action := "taint"
		if !tainted {
			action = "untaint"
		}
		return nil, fmt.Errorf("unable to %s runner: %w", action, res.Error)
	}

	runner.Tainted = tainted

	// Trigger leader election so the group picks a new leader after taint changes.
	if runner.RunnerGroupID != "" {
		s.evClient.Send(ctx, runner.RunnerGroupID, &runnergroupssignals.Signal{
			Type: runnergroupssignals.OperationElectLeader,
		})
	}

	return &runner, nil
}

// @ID						TaintRunner
// @Summary				taint a runner to exclude it from leader election
// @Tags					runners
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Param					runner_id	path	string	true	"runner ID"
// @Success				200	{object}	app.Runner
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/runners/{runner_id}/taint [post]
func (s *service) TaintRunner(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runner, err := s.setRunnerTainted(ctx, ctx.Param("runner_id"), true, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runner)
}

// @ID						UntaintRunner
// @Summary				untaint a runner to include it in leader election
// @Tags					runners
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Param					runner_id	path	string	true	"runner ID"
// @Success				200	{object}	app.Runner
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/runners/{runner_id}/untaint [post]
func (s *service) UntaintRunner(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runner, err := s.setRunnerTainted(ctx, ctx.Param("runner_id"), false, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runner)
}
