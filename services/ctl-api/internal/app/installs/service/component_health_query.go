package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// requireComponentHealthFeature gates the component-health read endpoints the
// same way GetInstallResources does.
func (s *service) requireComponentHealthFeature(ctx context.Context, org *app.Org) error {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth)
	if err != nil {
		return fmt.Errorf("unable to check component-health feature: %w", err)
	}
	if !enabled {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("component health is not enabled for org %s", org.ID),
			Description: "The component health feature is not enabled for this organization.",
		}
	}
	return nil
}

// findInstallComponent resolves :component_id to the install-component that
// health rows are keyed by. Accepts a catalog component ID first, falling
// back to the install-component's own ID, so either identifier works.
func (s *service) findInstallComponent(ctx context.Context, orgID, installID, componentID string) (*app.InstallComponent, error) {
	var ic app.InstallComponent
	err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{
			ComponentID: componentID,
			InstallID:   installID,
			OrgID:       orgID,
		}).
		First(&ic).Error
	if err == nil {
		return &ic, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{
			ID:        componentID,
			InstallID: installID,
			OrgID:     orgID,
		}).
		First(&ic).Error; err != nil {
		return nil, err
	}
	return &ic, nil
}

func (s *service) listHealthTransitions(ctx context.Context, orgID, installID, installComponentID string, from, to time.Time) ([]app.InstallComponentHealthTransition, error) {
	transitions := make([]app.InstallComponentHealthTransition, 0)
	if err := s.chDB.WithContext(ctx).
		Where(app.InstallComponentHealthTransition{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: installComponentID,
		}).
		Where("observed_at >= ? AND observed_at < ?", from, to).
		Order("observed_at ASC").
		Find(&transitions).Error; err != nil {
		return nil, fmt.Errorf("unable to query health transitions: %w", err)
	}
	return transitions, nil
}

// healthAtWindowStart returns the verdict in effect at `from` — the ToHealth
// of the latest transition before the window. Without it, a window starting
// after the component's last transition would read as unknown for days.
func (s *service) healthAtWindowStart(ctx context.Context, orgID, installID, installComponentID string, from time.Time) (string, error) {
	var seed app.InstallComponentHealthTransition
	err := s.chDB.WithContext(ctx).
		Where(app.InstallComponentHealthTransition{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: installComponentID,
		}).
		Where("observed_at < ?", from).
		Order("observed_at DESC").
		Limit(1).
		Find(&seed).Error
	if err != nil {
		return healthUnknown, fmt.Errorf("unable to query seed transition: %w", err)
	}
	if seed.ToHealth == "" {
		return healthUnknown, nil
	}
	return seed.ToHealth, nil
}

// findLatestBadTransition returns the most recent transition into degraded
// or unhealthy, whether or not it has since recovered. Returns (nil, nil)
// when there's no such transition.
func (s *service) findLatestBadTransition(ctx context.Context, orgID, installID, installComponentID string) (*app.InstallComponentHealthTransition, error) {
	rows := make([]app.InstallComponentHealthTransition, 0, 1)
	if err := s.chDB.WithContext(ctx).
		Where(app.InstallComponentHealthTransition{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: installComponentID,
		}).
		Where("to_health IN ?", []string{
			string(app.InstallComponentHealthStatusDegraded),
			string(app.InstallComponentHealthStatusUnhealthy),
		}).
		Order("observed_at DESC").
		Limit(1).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to query latest bad transition: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (s *service) nonHealthyResources(ctx context.Context, orgID, installID, installComponentID string) ([]app.InstallComponentResourceState, error) {
	resources := make([]app.InstallComponentResourceState, 0)
	if err := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(views.CurrentViewName(s.chDB, &app.InstallComponentResourceState{}))).
		Where(app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: installComponentID,
		}).
		Where("health != ?", string(app.InstallComponentHealthStatusHealthy)).
		Order("kind, namespace, name").
		Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("unable to query non-healthy resources: %w", err)
	}
	return resources, nil
}
