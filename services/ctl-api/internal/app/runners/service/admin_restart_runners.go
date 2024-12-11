package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"go.uber.org/zap"
)

type AdminRestartRunnersRequest struct {
}

type AdminRestartRunnersResponse struct {
	OrgID    string `json:"org_id"`
	RunnerID string `json:"runner_id"`
}

// @ID AdminRestartRunners
// @Summary	Restarts all non sandbox org and install runners
// @Description.markdown restart_runners.md
// @Param		req				body	AdminRestartRunnersRequest	true "Input"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array} AdminRestartRunnersResponse
// @Router		/v1/runners/restart [POST]
func (s *service) AdminRestartRunners(ctx *gin.Context) {
	var req AdminRestartRunnersRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	updatesResponse, err := s.bulkRestartRunners(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to restart runners: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, updatesResponse)
}

func (s *service) bulkRestartRunners(ctx context.Context) ([]AdminRestartRunnersResponse, error) {
	updatesResponse := []AdminRestartRunnersResponse{}
	batchSize := 50
	var runners []app.Runner
	offset := 0

	for {
		result := s.db.
			Joins("JOIN orgs ON runners.org_id = orgs.id AND orgs.org_type = ?", app.OrgTypeDefault).
			Offset(offset).
			Limit(batchSize).
			Find(&runners).
			Order("created_at ASC")

		if result.Error != nil {
			return nil, fmt.Errorf("unable to fetch runners: %w", result.Error)
		}

		if len(runners) == 0 {
			break
		}

		for _, runner := range runners {
			job, err := s.adminCreateJob(ctx, runner.ID, app.RunnerJobTypeShutDown)
			if err != nil {
				s.l.Error("unable to create shutdown job", zap.String("runner_id", runner.ID), zap.Error(err))
			} else {
				updatesResponse = append(updatesResponse, AdminRestartRunnersResponse{
					OrgID:    runner.OrgID,
					RunnerID: runner.ID,
				})
			}

			s.evClient.Send(ctx, runner.ID, &signals.Signal{
				Type:  signals.OperationProcessJob,
				JobID: job.ID,
			})
		}

		offset += batchSize
	}

	return updatesResponse, nil
}
