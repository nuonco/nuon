package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type AdminBulkUpdateRunnersRequest struct {
	ContainerImageTag string `json:"container_image_tag"`
}

type AdminBulkUpdateRunnersResponse struct {
	OrgID         string `json:"org_id"`
	RunnerGroupID string `json:"runner_group_id"`
}

//	@ID						AdminBulkUpdateRunners
//	@Summary				Admin Bulk Update Runners
//	@Description.markdown	admin_bulk_update_runners.md
//	@Param					req	body	AdminBulkUpdateRunnersRequest	true	"Input"
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{array}	AdminBulkUpdateRunnersResponse
//	@Router					/v1/runners/bulk-update [PATCH]
func (s *service) AdminBulkUpdateRunners(ctx *gin.Context) {
	var req AdminBulkUpdateRunnersRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	updatesResponse, err := s.bulkUpdateRunners(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, updatesResponse)
}

func (s *service) bulkUpdateRunners(ctx context.Context, req *AdminBulkUpdateRunnersRequest) ([]AdminBulkUpdateRunnersResponse, error) {
	updatesResponse := []AdminBulkUpdateRunnersResponse{}
	batchSize := 50
	var runnerGroups []app.RunnerGroup
	offset := 0

	for {
		result := s.db.
			Joins("JOIN orgs ON runner_groups.org_id = orgs.id AND orgs.org_type = ?", app.OrgTypeDefault).
			Where("type = ?", app.RunnerGroupTypeOrg).
			Offset(offset).
			Limit(batchSize).
			Find(&runnerGroups).
			Order("created_at ASC")

		if result.Error != nil {
			return nil, fmt.Errorf("unable to fetch runner groups: %w", result.Error)
		}

		if len(runnerGroups) == 0 {
			break
		}

		for _, runnerGroup := range runnerGroups {
			updates := app.RunnerGroupSettings{
				ContainerImageTag: req.ContainerImageTag,
			}
			obj := app.RunnerGroupSettings{
				RunnerGroupID: runnerGroup.ID,
			}

			res := s.db.WithContext(ctx).
				Where(obj).
				Updates(updates)

			if res.Error != nil {
				return nil, fmt.Errorf("unable to update runner group settings: %w", res.Error)
			}

			updatesResponse = append(updatesResponse, AdminBulkUpdateRunnersResponse{
				OrgID:         runnerGroup.OrgID,
				RunnerGroupID: runnerGroup.ID,
			})
		}

		offset += batchSize
	}

	orgVisited := make(map[string]bool)
	for _, response := range updatesResponse {
		if _, ok := orgVisited[response.OrgID]; !ok {
			ctx = cctx.SetOrgIDContext(ctx, response.OrgID)
			s.evClient.Send(ctx, response.OrgID, &signals.Signal{
				Type: signals.OperationReprovision,
			})
			orgVisited[response.OrgID] = true
		}
	}

	return updatesResponse, nil
}
