package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AllOrgsResponse []*app.Org

// @BasePath /v1/orgs

// Return all orgs
// @Summary Return all orgs
// @Schemes
// @Description return all orgs
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} app.Org
// @Router /v1/orgs [GET]
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
	res := s.db.WithContext(ctx).Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all orgs: %w", res.Error)
	}

	return orgs, nil
}
