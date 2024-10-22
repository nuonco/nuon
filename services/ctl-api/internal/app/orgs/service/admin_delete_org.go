package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	appsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	installsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminDeleteOrgRequest struct {
	Force bool `json:"force"`
}

// @ID AdminDeleteOrg
// @Summary delete an org and everything in it
// @Description.markdown delete_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminDeleteOrgRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-delete [POST]
func (s *service) AdminDeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminDeleteOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if org.OrgType == app.OrgTypeIntegration {
		err := s.deleteIntegrationOrg(ctx, org.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, true)
		return
	}

	// regular delete for orgs
	org, err = s.getOrgAndDependencies(ctx, org.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org dependencies"))
		return
	}
	orgID = org.ID

	// restart the org, to ensure the event loop is active
	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationRestart,
	})

	// delete the org
	err = s.deleteOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// TODO(jm): this should happen in the event loop
	// send a signal to all children
	for _, app := range org.Apps {
		s.evClient.Send(ctx, app.ID, &appsignals.Signal{
			Type: appsignals.OperationDeleted,
		})
		s.evClient.Send(ctx, app.ID, &appsignals.Signal{
			Type: appsignals.OperationDeprovision,
		})

		for _, install := range app.Installs {
			s.evClient.Send(ctx, install.ID, &installsignals.Signal{
				Type: installsignals.OperationDelete,
			})
			s.evClient.Send(ctx, install.ID, &installsignals.Signal{
				Type: installsignals.OperationForgotten,
			})
		}

		for _, comp := range app.Components {
			s.evClient.Send(ctx, comp.ID, &componentsignals.Signal{
				Type: componentsignals.OperationDelete,
			})
		}
	}
	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationDeprovision,
	})
	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type: sigs.OperationDelete,
	})

	if req.Force {
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigs.OperationForceDelete,
		})
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getOrgAndDependencies(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Installs.RunnerGroup").
		Preload("Apps.Installs.RunnerGroup.Runners").
		Preload("Apps.Components").
		Where("name = ?", orgID).
		Or("id = ?", orgID).
		First(&org)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}
