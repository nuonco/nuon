package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

type AdminDeleteCanaryOrgsRequest struct {
	Force bool `json:"force"`
}

//	@ID						AdminDeleteCanaryOrgs
//	@Summary				delete canary orgs
//	@Description.markdown	delete_org.md
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminDeleteCanaryOrgsRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/admin-delete-canarys [POST]
func (s *service) AdminDeleteCanaryOrgs(ctx *gin.Context) {
	orgs, err := s.getCanaryOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, org := range orgs {
		s.evClient.Send(ctx, org.ID, &signals.Signal{
			Type: signals.OperationDelete,
		})
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getCanaryOrgs(ctx context.Context) ([]app.Org, error) {
	var orgs []app.Org
	res := s.db.WithContext(ctx).
		Joins("JOIN accounts on orgs.created_by_id=accounts.id").
		Where("accounts.account_type = ?", "canary").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get canary orgs: %w", res.Error)
	}

	return orgs, nil
}
