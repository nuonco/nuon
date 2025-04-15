package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallWorkflow
// @Summary				get an install workflow
// @Description.markdown	get_install_workflow.md
// @Param				install_workflow_id path	string	true	"install workflow ID"
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
// @Success				200	{object}		app.InstallWorkflow
// @Router					/v1/install-workflows/{install_workflow_id} [GET]
func (s *service) GetInstallWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")

	installWorkflow, err := s.getInstallWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflows"))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflow)
}

func (s *service) getInstallWorkflow(ctx *gin.Context, workflowID string) (*app.InstallWorkflow, error) {
	var installWorkflow app.InstallWorkflow
	res := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("idx ASC")
		}).
		Where(app.InstallWorkflow{
			ID: workflowID,
		}).
		First(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow")
	}

	return &installWorkflow, nil
}
