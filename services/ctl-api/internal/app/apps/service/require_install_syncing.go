package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

func (s *service) requireInstallSyncing(ctx *gin.Context) bool {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppInstallSyncing)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return false
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppInstallSyncing))
		return false
	}

	return true
}
