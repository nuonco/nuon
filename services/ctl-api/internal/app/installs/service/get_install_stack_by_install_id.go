package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

// @ID						GetInstallStackByInstallID
// @Summary				get an install stack by install ID
// @Description.markdown	get_install_stack.md
// @Param					install_id					path	string	true	"install ID"
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
// @Success				200	{array}		app.InstallStack
// @Router					/v1/installs/{install_id}/stack [GET]
func (s *service) GetInstallStackByInstallID(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("install_id")
	installStack, err := s.getInstallStack(ctx, appID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	if installStack == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "install stack not found"})
		return
	}

	ctx.JSON(http.StatusOK, installStack)
}

func (s *service) getInstallStack(ctx *gin.Context, installID, orgID string) (*app.InstallStack, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("InstallStack").
		Preload("InstallStack.InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC").Limit(10)
		}).
		Preload("InstallStack.InstallStackVersions.Runs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_version_runs.created_at DESC").Limit(10)
		}).
		Where("id = ? and org_id = ?", installID, orgID).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return install.InstallStack, nil
}
