package service

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminShutdownAllRunnerProcessesRequest struct {
	ShutdownType *app.RunnerProcessShutdownType `json:"shutdown_type,omitempty"`
	ProcessType  *app.RunnerProcessType         `json:"process_type,omitempty"`
}

type AdminShutdownAllRunnerProcessesResponse struct {
	ProcessesShutdown int `json:"processes_shutdown"`
	ProcessCount      int `json:"process_count"`
	LatestCount       int `json:"latest_count"`
}

// @ID						AdminShutdownAllRunnerProcesses
// @Summary				Shutdown all active runner processes across all orgs
// @Param					req	body	AdminShutdownAllRunnerProcessesRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminShutdownAllRunnerProcessesResponse
// @Router					/v1/runners/shutdown-processes [POST]
func (s *service) AdminShutdownAllRunnerProcesses(ctx *gin.Context) {
	var req AdminShutdownAllRunnerProcessesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	shutdownType := app.RunnerProcessShutdownTypeGraceful
	if req.ShutdownType != nil {
		shutdownType = *req.ShutdownType
	}

	resp, err := s.shutdownAllRunnerProcesses(ctx, shutdownType, req.ProcessType)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) shutdownAllRunnerProcesses(ctx context.Context, shutdownType app.RunnerProcessShutdownType, processType *app.RunnerProcessType) (*AdminShutdownAllRunnerProcessesResponse, error) {
	var allProcesses []app.RunnerProcess
	query := s.db.WithContext(ctx).
		Where("runner_processes.composite_status::jsonb ->> 'status' IN ('active', 'offline')").
		Order("runner_processes.runner_id, runner_processes.type, runner_processes.created_at DESC")

	if processType != nil {
		query = query.Where("runner_processes.type = ?", *processType)
	}

	res := query.Find(&allProcesses)
	if res.Error != nil {
		return nil, res.Error
	}

	// Deduplicate: keep only the most recent process per runner+type.
	type key struct {
		RunnerID string
		Type     app.RunnerProcessType
	}
	seen := make(map[key]bool)
	var latest []app.RunnerProcess

	for _, p := range allProcesses {
		k := key{RunnerID: p.RunnerID, Type: p.Type}
		if !seen[k] {
			seen[k] = true
			latest = append(latest, p)
		}
	}

	// Build all shutdown records and batch-insert them.
	shutdowns := make([]app.RunnerProcessShutdown, len(latest))
	for i, p := range latest {
		shutdowns[i] = app.RunnerProcessShutdown{
			RunnerProcessID: p.ID,
			OrgID:           p.OrgID,
			CreatedByID:     p.CreatedByID,
			Type:            shutdownType,
			CompositeStatus: app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusRequested)),
		}
	}

	created := 0
	if len(shutdowns) > 0 {
		if res := s.db.WithContext(ctx).Create(&shutdowns); res.Error != nil {
			return nil, res.Error
		}
		created = len(shutdowns)
	}

	return &AdminShutdownAllRunnerProcessesResponse{
		ProcessesShutdown: created,
		ProcessCount:      len(allProcesses),
		LatestCount:       len(latest),
	}, nil
}
