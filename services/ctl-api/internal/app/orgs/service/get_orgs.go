package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
)

// @Summary	Return current user's orgs
// @Description.markdown get_orgs.md
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Org
// @Router			/v1/orgs [GET]
func (s *service) GetCurrentUserOrgs(ctx *gin.Context) {
	userToken, err := auth.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	orgs, err := s.getCurrentUserOrgs(ctx, userToken.Subject)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getCurrentUserOrgs(ctx context.Context, userID string) ([]*app.Org, error) {
	var userOrgs []*app.UserOrg

	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		Where(&app.UserOrg{
			UserID: userID,
		}).Find(&userOrgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get current user's orgs: %w", res.Error)
	}

	var orgs []*app.Org
	for _, userOrg := range userOrgs {
		orgs = append(orgs, &userOrg.Org)
	}

	return orgs, nil
}
