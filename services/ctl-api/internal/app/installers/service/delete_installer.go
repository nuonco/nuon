package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID DeleteInstaller
// @Summary	delete an installer
// @Description.markdown	delete_installer.md
// @Param			installer_id	path	string	true	"installer ID"
// @Tags installers
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{boolean}	true
// @Router			/v1/installers/{installer_id} [DELETE]
func (s *service) DeleteInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")

	err := s.deleteInstaller(ctx, installerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteInstaller(ctx context.Context, appInstallerID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Installer{
		ID: appInstallerID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete installer: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("installer not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}
