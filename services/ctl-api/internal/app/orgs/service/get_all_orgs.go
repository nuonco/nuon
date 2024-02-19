package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID AdminGetAllOrgs
// @BasePath	/v1/orgs
// @Summary	Return all orgs
// @Schemes
// @Description	return all orgs
// @Tags			orgs/admin
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.Org
// @Router			/v1/orgs [GET]
func (s *service) GetAllOrgs(ctx *gin.Context) {
	orgs, err := s.getAllOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getAllOrgs(ctx context.Context) ([]*app.Org, error) {
	var orgs []*app.Org
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Joins("JOIN user_tokens ON user_tokens.subject=orgs.created_by_id").
		Where("user_tokens.token_type = ?", app.TokenTypeAuth0).
		Order("orgs.created_at desc").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all orgs: %w", res.Error)
	}

	return orgs, nil
}
