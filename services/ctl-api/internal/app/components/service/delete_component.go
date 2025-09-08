package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

// @ID						DeleteComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/components/{component_id} [DELETE]
func (s *service) DeleteComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	err := s.deleteComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteComponent(ctx context.Context, compID string) error {
	comp := app.Component{
		ID: compID,
	}

	res := s.db.WithContext(ctx).Model(&comp).First(&comp).Where("id = ?", compID)
	if res.Error != nil {
		return fmt.Errorf("unable to get component %s: %w", compID, res.Error)
	}

	dependentComponents, err := s.appsHelpers.GetComponentDependents(ctx, comp.AppID, comp.ID)
	if err != nil {
		return fmt.Errorf("unable to get dependents: %w", err)
	}

	if len(dependentComponents) > 0 {
		componentsIds := make([]string, 0, len(dependentComponents))
		for _, c := range dependentComponents {
			componentsIds = append(componentsIds, c)
		}
		return fmt.Errorf("unable to delete component %s, components dependents exist Dependent IDs: %s", compID, componentsIds)
	}

	res = s.db.WithContext(ctx).Model(&comp).Updates(app.Component{
		Status:            "delete_queued",
		StatusDescription: "delete has been queued and waiting",
	})

	if res.Error != nil {
		return fmt.Errorf("unable to update component: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("component not found %s: %w", compID, gorm.ErrRecordNotFound)
	}

	return nil
}
