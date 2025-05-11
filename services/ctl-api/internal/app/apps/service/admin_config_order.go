package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type AdminAppConfigOrderRequest struct {
	ConfigID string
}

type AdminAppConfigOrderResult struct {
	Name   string
	CompID string
}

// @ID						AdminAppConfigOrderApp
// @Summary				get an app's graph
// @Description.markdown  app_config_component_order.md
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Param			req		body	AdminAppConfigOrderRequest	true	"Input"
// @Param				app_id	path	string					true	"app id"
// @Produce				json
// @Success				201	{array}	AdminAppConfigOrderResult
// @Router					/v1/apps/{app_id}/admin-config-order [POST]
func (s *service) AdminConfigOrder(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminAppConfigOrderRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	cfgID := req.ConfigID
	if cfgID == "" {
		cfgs, err := s.getAppConfigs(ctx, orgID, appID)
		if err != nil {
			ctx.Error(err)
			return
		}
		cfgID = cfgs[0].ID
	}

	appCfg, err := s.helpers.GetFullAppConfig(ctx, cfgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	compIDs, err := s.helpers.GetConfigDefaultComponentOrder(ctx, appCfg)
	if err != nil {
		ctx.Error(err)
		return
	}

	comps, err := s.helpers.GetAppComponentsAndLatestConfigConnection(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	output := make([]AdminAppConfigOrderResult, 0)
	for _, compID := range compIDs {
		found := false
		for _, comp := range comps {
			if comp.ID != compID {
				continue
			}

			found = true
			output = append(output, AdminAppConfigOrderResult{
				Name:   comp.Name,
				CompID: comp.ID,
			})
			break
		}

		if !found {
			ctx.Error(errors.New("component was not found"))
			return
		}
	}

	ctx.JSON(http.StatusOK, output)
	return
}
