package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"gorm.io/gorm"
)

// @ID DeleteOrg
// @Summary	Delete an org
// @Schemes
// @Description.markdown	delete_org.md
// @Tags			orgs
// @Accept			json
// @Security APIKey
// @Security OrgID
// @Produce		json
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{boolean}	ok
// @Router			/v1/orgs/current [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.deleteOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, org.ID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteOrg(ctx context.Context, orgID string) error {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		StatusDescription: "delete has been queued",
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found: %w", gorm.ErrRecordNotFound)
	}
	return nil
}
