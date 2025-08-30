package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

// @ID 					 	DeleteVCSConnection
// @Summary 			 	Deletes a VCS connection
// @Description.markdown 	delete_vcs_connection.md
// @Param connection_id  	path string true "Connection ID"
// @Tags 				 	vcs
// @Accept 				 	json
// @Produce 			 	json
// @Security			 	APIKey
// @Security			 	OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					204
// @Router 					/v1/vcs/connections/{connection_id} [delete]
func (s *service) DeleteConnection(ctx *gin.Context) {
	vcsID := ctx.Param("connection_id")

	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, vcsID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connection: %w", err))
		return
	}

	err = s.deleteGithubAppInstallation(ctx, vcsConn.GithubInstallID)
	if err != nil {
		// If we can't delete the Github App installation, we should still try
		// to delete the connection from our DB.
		s.l.Info(err.Error())
	}

	err = s.deleteConnection(ctx, vcsConn)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusNoContent, true)
}

func (s *service) deleteConnection(ctx context.Context, vcsConn *app.VCSConnection) error {
	res := s.db.WithContext(ctx).Delete(vcsConn)
	if res.Error != nil {
		return fmt.Errorf("unable to delete connection: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("connection not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *service) deleteGithubAppInstallation(ctx context.Context, installId string) error {
	ghClient, err := s.helpers.GetJWTVCSConnectionClient()

	iInstallId, err := strconv.ParseInt(installId, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert github install ID to int: %w", err)
	}

	_, err = ghClient.Apps.DeleteInstallation(ctx, iInstallId)
	if err != nil {
		return fmt.Errorf("unable to delete github app installation id: %d: %w", iInstallId, err)
	}

	return nil
}
