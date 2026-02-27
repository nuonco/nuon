package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminDeleteOrgRequest struct {
	Force bool `json:"force"`
}

// @ID						AdminDeleteOrg
// @Summary				delete an org and everything in it
// @Description.markdown	delete_org.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminDeleteOrgRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-delete [POST]
func (s *service) AdminDeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminDeleteOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// ShouldBindJSON returns error for invalid JSON but accepts empty body
		// For empty body, Force defaults to false which is the desired behavior
		if err.Error() != "EOF" {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("unable to parse request: %w", err),
				Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
			})
			return
		}
	}

	org, err := s.adminGetOrg(ctx, orgID)
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
		ForceDelete: req.Force,
	})

	ctx.JSON(http.StatusOK, true)
}
