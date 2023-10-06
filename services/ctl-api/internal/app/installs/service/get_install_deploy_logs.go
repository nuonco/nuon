package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

type DeployLog struct{}

//	@BasePath	/v1/installs
//
// Get install deploy logs
//
//	@Summary	get install deploy logs
//	@Schemes
//	@Description	get install deploy logs
//	@Param			install_id	path	string	true	"install ID"
//	@Param			deploy_id	path	string	true	"deploy ID"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	[]DeployLog
//	@Router			/v1/installs/{install_id}/deploys/{deploy_id}/logs [get]
func (s *service) GetInstallDeployLogs(ctx *gin.Context) {
	ctx.Error(fmt.Errorf("not yet implemented"))
}

func (s *service) getInstallDeployLogs(ctx context.Context, installID, componentID string) ([]DeployLog, error) {
	return nil, nil
}
