package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID 					 	DeleteVCSConnection
// @Summary 			 	Deletes a VCS connection
// @Description.markdown 	delete_vcs_connection.md
// @Param connection_id  	path	string	true	"Connection ID"
// @Param delete_github_app	query	bool	false	"If true, also uninstall the GitHub App on the GitHub side. Defaults to false so other Nuon orgs sharing the same installation are not impacted."
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

	deleteGithubApp := false
	if raw := ctx.Query("delete_github_app"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			ctx.Error(fmt.Errorf("invalid delete_github_app value: %w", err))
			return
		}
		deleteGithubApp = parsed
	}

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

	if deleteGithubApp {
		if err := s.ghClient.DeleteInstallation(ctx, vcsConn.GithubInstallID); err != nil {
			// If we can't delete the Github App installation, we should still
			// try to delete the connection from our DB.
			s.l.Info(err.Error())
		}
	}

	// Stop all queues owned by this VCS connection
	for _, q := range vcsConn.Queues {
		if err := s.helpers.StopConnectionQueue(ctx, q.ID); err != nil {
			s.l.Warn("unable to stop vcs connection queue",
				zap.String("vcs_connection_id", vcsConn.ID),
				zap.String("queue_id", q.ID),
				zap.Error(err),
			)
		}
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
