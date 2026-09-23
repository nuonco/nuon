package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appbranchchanged"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type MoveInstallToAppBranchRequest struct {
	// AppBranchID is the branch to move the install to. It must belong to the
	// install's app and have an app config to deploy.
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

func (r *MoveInstallToAppBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						MoveInstallToAppBranch
// @Summary				move an install to another app branch
// @Description			Moves the install to the given app branch and reconciles it onto that branch's current app config. An install belongs to exactly one app branch and this is the only way to change which one; labels and install group selectors decide which group inside the owning branch deploys it. The destination branch must belong to the same app and have an active, non-preview app config. There is no way to move an install off a branch without naming another.
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	MoveInstallToAppBranchRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Install
// @Router					/v1/installs/{install_id}/app-branch [PATCH]
func (s *service) MoveInstallToAppBranch(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	installID := ctx.Param("install_id")

	var req MoveInstallToAppBranchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).First(&branch, "id = ?", req.AppBranchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("app branch %s not found", req.AppBranchID),
				Description: "The selected app branch does not exist.",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to get app branch %s: %w", req.AppBranchID, err))
		return
	}

	if branch.AppID != install.AppID {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("app branch %s belongs to app %s, install %s to app %s", branch.ID, branch.AppID, install.ID, install.AppID),
			Description: "The selected app branch belongs to a different app.",
		})
		return
	}

	if install.AppBranchID.Valid && install.AppBranchID.String == branch.ID {
		ctx.JSON(http.StatusOK, install)
		return
	}

	// Refusing up front beats moving the install somewhere that has nothing to
	// deploy and leaving it stranded there.
	target, err := s.helpers.ResolveAppBranchRunForInstall(ctx, branch.ID, &install)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.appsHelpers.SetInstallAppBranch(ctx, install.ID, branch.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to move install to app branch: %w", err))
		return
	}

	if err := appbranchchanged.Enqueue(ctx, s.queueClient, install.ID, branch.ID, target.InstallGroupID); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue app branch install update: %w", err))
		return
	}

	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to reload install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}
