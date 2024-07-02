package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	appsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	installsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminDeleteCanaryOrgsRequest struct{}

// @ID AdminDeleteCanaryOrgs
// @Summary delete canary orgs
// @Description.markdown delete_org.md
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminDeleteCanaryOrgsRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/admin-delete-canarys [POST]
func (s *service) AdminDeleteCanaryOrgs(ctx *gin.Context) {
	orgs, err := s.getCanaryOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, org := range orgs {
		for _, app := range org.Apps {
			for _, install := range app.Installs {
				s.evClient.Send(ctx, install.ID, &installsignals.Signal{
					Type: installsignals.OperationForgotten,
				})
			}

			for _, component := range app.Components {
				s.evClient.Send(ctx, component.ID, &componentsignals.Signal{
					Type: componentsignals.OperationDelete,
				})
			}

			s.evClient.Send(ctx, app.ID, &appsignals.Signal{
				Type: appsignals.OperationDeleted,
			})
			s.evClient.Send(ctx, app.ID, &appsignals.Signal{
				Type: appsignals.OperationDeprovision,
			})
		}
		s.evClient.Send(ctx, org.ID, &signals.Signal{
			Type: signals.OperationForceDelete,
		})
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getCanaryOrgs(ctx context.Context) ([]app.Org, error) {
	var orgs []app.Org
	res := s.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Components").
		Joins("JOIN accounts on orgs.created_by_id=accounts.id").
		Where("accounts.account_type = ?", "canary").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get canary orgs: %w", res.Error)
	}

	return orgs, nil
}
