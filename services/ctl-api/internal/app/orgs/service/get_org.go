package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *service) GetOrg(ctx *gin.Context) {
	orgID := ctx.Param("id")
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) getOrg(ctx context.Context, orgID string) ([]*app.Org, error) {
	var orgs []*app.Org
	res := s.db.WithContext(ctx).Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all orgs: %w", res.Error)
	}

	return orgs, nil
}
