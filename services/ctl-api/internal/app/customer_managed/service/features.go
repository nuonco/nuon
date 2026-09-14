package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

func (s *service) customerManagedInstallsEnabled(ctx *gin.Context) bool {
	if s.features == nil {
		ctx.Error(fmt.Errorf("customer-managed installs feature client is unavailable"))
		return false
	}
	enabled, err := s.features.FeatureEnabled(ctx, app.OrgFeatureCustomerManagedInstalls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check customer-managed installs feature: %w", err))
		return false
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureCustomerManagedInstalls))
		return false
	}
	return true
}
