package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

// @ID						GetInstallStack
// @Summary				get an install stack by stack ID
// @Description.markdown	get_install_stack.md
// @Param					stack_id					path	string	true	"stack ID"
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
// @Success				200	{object}		app.InstallStack
// @Router					/v1/installs/stacks/{stack_id} [GET]
func (s *service) GetInstallStackByStackID(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("stack_id")
	installStack, err := s.getInstallStackByStackID(ctx, appID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installStack)
}

func (s *service) getInstallStackByStackID(ctx *gin.Context, installStackID, orgID string) (*app.InstallStack, error) {
	installStack := &app.InstallStack{}
	res := s.db.WithContext(ctx).
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC").Limit(10)
		}).
		Preload("InstallStackVersions.Runs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_version_runs.created_at DESC").Limit(10)
		}).
		Preload("InstallStackOutputs").
		Where("id = ? and org_id = ?", installStackID, orgID).
		First(&installStack, "id = ?", installStackID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return installStack, nil
}
