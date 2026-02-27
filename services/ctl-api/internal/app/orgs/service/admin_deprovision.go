package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminDeprovisionOrgRequest struct {
	Force bool `json:"force"`
}

// @ID						AdminDeprovisionOrg
// @Summary				deprovision an org, but keep it in the database
// @Description.markdown	deprovision_org.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminDeprovisionOrgRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")
	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminDeprovisionOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// Validate that all apps have been deprovisioned before allowing org deprovision
	// Force flag does NOT bypass this check - apps must always be deleted first
	var orgWithApps app.Org
	if err := s.db.WithContext(ctx).Preload("Apps").First(&orgWithApps, "id = ?", org.ID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to check org apps: %w", err))
		return
	}

	if len(orgWithApps.Apps) > 0 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("cannot deprovision org with active apps"),
			Description: fmt.Sprintf("organization has %d app(s) that must be deleted before the organization can be deprovisioned", len(orgWithApps.Apps)),
		})
		return
	}

	sigTyp := sigs.OperationDeprovision
	if req.Force {
		sigTyp = sigs.OperationForceDeprovision
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigTyp,
	})

	ctx.JSON(http.StatusOK, true)
}
