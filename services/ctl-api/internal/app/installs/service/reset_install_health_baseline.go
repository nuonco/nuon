package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// healthBaselineMetadataKey stores the install's health-window baseline
// (RFC3339) in install metadata. Timeline/uptime reads clamp their window
// start to it, so bring-up churn stops counting without rewriting history.
const healthBaselineMetadataKey = "health_baseline_at"

type ResetInstallHealthBaselineResponse struct {
	BaselineAt time.Time `json:"baseline_at"`
}

// @ID						ResetInstallHealthBaseline
// @Summary				reset the install's health window
// @Description			Sets the install's health baseline to now: uptime and the health timeline start counting from this moment. Past observations stay recorded but no longer count toward uptime. Requires the component-health feature.
// @Param					install_id	path	string	true	"install ID"
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
// @Success				200	{object}	service.ResetInstallHealthBaselineResponse
// @Router					/v1/installs/{install_id}/health/baseline [post]
func (s *service) ResetInstallHealthBaseline(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireComponentHealthFeature(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	now := time.Now().UTC()
	if err := s.setHealthBaseline(ctx, org.ID, installID, now); err != nil {
		ctx.Error(fmt.Errorf("unable to reset health baseline: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ResetInstallHealthBaselineResponse{BaselineAt: now})
}

func (s *service) setHealthBaseline(ctx context.Context, orgID, installID string, at time.Time) error {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if install.Metadata == nil {
		install.Metadata = map[string]*string{}
	}
	val := at.Format(time.RFC3339)
	install.Metadata[healthBaselineMetadataKey] = &val

	if err := s.db.WithContext(ctx).
		Model(&app.Install{ID: install.ID}).
		Update("metadata", install.Metadata).Error; err != nil {
		return fmt.Errorf("unable to update install metadata: %w", err)
	}
	return nil
}

// healthBaseline returns the install's baseline, zero when never reset.
func (s *service) healthBaseline(ctx context.Context, orgID, installID string) (time.Time, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Select("id", "metadata").
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return time.Time{}, fmt.Errorf("unable to get install: %w", err)
	}
	raw, ok := install.Metadata[healthBaselineMetadataKey]
	if !ok || raw == nil {
		return time.Time{}, nil
	}
	at, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return time.Time{}, nil
	}
	return at, nil
}

// clampToBaseline moves the window start forward to the baseline when one is
// set inside the window.
func clampToBaseline(from, baseline time.Time) time.Time {
	if !baseline.IsZero() && baseline.After(from) {
		return baseline
	}
	return from
}
