package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallStackRuns
// @Summary				get an install's stack runs
// @Description	get install stack runs
// @Param install_id					path	string	true "install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}		app.InstallStackVersionRun
// @Router					/v1/installs/{install_id}/stack-runs [get]
func (s *service) GetInstallStackRuns(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	runs, err := s.getInstallLatestStackRunsByStackID(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runs)
}

func (s *service) getInstallLatestStackRunsByStackID(ctx context.Context, installID string) ([]app.InstallStackVersionRun, error) {
	var runs []app.InstallStackVersionRun

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).

		// join component-releases to component-builds to component-config-connections to components
		Joins("JOIN install_stack_versions ON install_stack_versions.id=install_stack_version_runs.install_stack_version_id").
		Joins("JOIN install_stacks ON install_stacks.id=install_stack_versions.install_stack_id").
		Where("install_stacks.install_id = ?", installID).
		Order("install_stack_version_runs.created_at DESC").
		Find(&runs)
	if res.Error == nil {
		// Pull in each run's log stream so the dashboard can render logs
		// without an extra round-trip. WriteToken/RunnerAPIURL stay zero —
		// they're transient response-only fields populated on POST.
		s.attachRunLogStreams(ctx, runs)
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component releases")
	}

	return runs, nil
}

// attachRunLogStreams hydrates each run's transient LogStream pointer with
// its persisted log_stream row in one batched query. Best-effort: a missing
// or unfetchable log stream just leaves LogStream nil.
func (s *service) attachRunLogStreams(ctx context.Context, runs []app.InstallStackVersionRun) {
	ids := make([]string, 0, len(runs))
	for _, r := range runs {
		if r.LogStreamID != "" {
			ids = append(ids, r.LogStreamID)
		}
	}
	if len(ids) == 0 {
		return
	}
	var streams []app.LogStream
	if res := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&streams); res.Error != nil {
		return
	}
	byID := make(map[string]*app.LogStream, len(streams))
	for i := range streams {
		byID[streams[i].ID] = &streams[i]
	}
	for i := range runs {
		if ls, ok := byID[runs[i].LogStreamID]; ok {
			runs[i].LogStream = ls
		}
	}
}
