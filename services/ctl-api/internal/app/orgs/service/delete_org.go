package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID			DeleteOrg
// @Summary	Delete an org
// @Schemes
// @Description.markdown	delete_org.md
// @Tags					orgs
// @Accept					json
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	ok
// @Router					/v1/orgs/current [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Validate that all apps have been deprovisioned before allowing org deletion
	var orgWithApps app.Org
	if err := s.db.WithContext(ctx).Preload("Apps").First(&orgWithApps, "id = ?", org.ID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to check org apps: %w", err))
		return
	}

	if len(orgWithApps.Apps) > 0 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("cannot delete org with active apps"),
			Description: fmt.Sprintf("organization has %d app(s) that must be deleted before the organization can be deleted", len(orgWithApps.Apps)),
		})
		return
	}

	if org.OrgType == app.OrgTypeIntegration {
		err := s.helpers.HardDelete(ctx, org.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, true)
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type:        sigs.OperationDelete,
		ForceDelete: false,
	})

	ctx.JSON(http.StatusOK, true)
}
