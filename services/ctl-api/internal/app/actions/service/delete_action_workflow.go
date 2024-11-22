package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID DeleteActionWorkflow
// @Summary	delete an app
// @Description.markdown	delete_action_workflow.md
// @Param			action_workflow_id path	string	true	"action workflow ID"
// @Tags			actions
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
// @Router			/v1/action-workflows/{action_workflow_id} [DELETE]
func (s *service) DeleteActionWorkflow(ctx *gin.Context) {
	awID := ctx.Param("action_workflow_id")

	err := s.deleteActionWorkflow(ctx, awID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteActionWorkflow(ctx context.Context, awID string) error {
	aw := app.ActionWorkflow{ID: awID}

	//TODO: we may want to queue up delete
	res := s.db.WithContext(ctx).Delete(&aw)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("unable to delete action workflow %s: %w", awID, res.Error)
	}
	return nil
}
