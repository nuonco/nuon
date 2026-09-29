package service

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID				DeleteRunbook
// @Summary		delete a runbook
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string	true	"app ID"
// @Param			runbook_id	path	string	true	"runbook ID"
// @Success		200			{object}	bool
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		409			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id} [delete]
func (s *service) DeleteRunbook(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	runbookID := ctx.Param("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	branchNames, err := s.branchesUsingRunbookAsPostDeploy(ctx, org.ID, appID, runbookID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if len(branchNames) > 0 {
		ctx.Error(stderr.ErrConflict{
			Err: fmt.Errorf("runbook %s is a post-deploy runbook on branch(es): %s", runbookID, strings.Join(branchNames, ", ")),
			Description: fmt.Sprintf(
				"this runbook runs after deploys on branch(es) %s. Remove it from those branches' post-deploy runbooks first, otherwise their next run would try to execute a deleted runbook.",
				strings.Join(branchNames, ", "),
			),
		})
		return
	}

	res := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: org.ID}).
		Delete(&app.Runbook{}, "id = ?", runbookID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete runbook: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) branchesUsingRunbookAsPostDeploy(ctx *gin.Context, orgID, appID, runbookID string) ([]string, error) {
	var branches []app.AppBranch
	if err := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: orgID, AppID: appID}).
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to list app branches: %w", err)
	}

	var names []string
	for _, branch := range branches {
		var config app.AppBranchConfig
		err := s.db.WithContext(ctx).
			Where(app.AppBranchConfig{AppBranchID: branch.ID}).
			Order("created_at DESC").
			First(&config).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get config for app branch %s: %w", branch.ID, err)
		}

		if slices.Contains(config.PostDeployRunbookIDs, runbookID) {
			name := branch.Name
			if name == "" {
				name = branch.ID
			}
			names = append(names, name)
		}
	}

	return names, nil
}
