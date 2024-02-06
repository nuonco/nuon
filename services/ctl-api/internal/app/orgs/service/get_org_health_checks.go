package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// @ID GetOrgHealthChecks
// @Summary	Get an org's health checks
// @Description.markdown	get_org_health_checks.md
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Param   limit  query int	 false	"limit of health checks to return"	     Default(60)
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}	app.OrgHealthCheck
// @Router			/v1/orgs/current/health-checks [GET]
func (s *service) GetOrgHealthChecks(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	limitStr := ctx.DefaultQuery("limit", "60")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	healthChecks, err := s.getOrgHealthChecks(ctx, org.ID, limitVal)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, healthChecks)
}

func (s *service) getOrgHealthChecks(ctx context.Context, orgID string, limit int) ([]app.OrgHealthCheck, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Preload("HealthChecks", func(db *gorm.DB) *gorm.DB {
			return db.Where("org_id = ?", orgID).Order("org_health_checks.created_at DESC").Limit(limit)
		}).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return org.HealthChecks, nil
}
