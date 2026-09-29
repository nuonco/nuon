package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

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
		Scopes(scopes.WithOverrideTable(app.InstallComponentResourceStatesLatestView)).
		Where(app.InstallComponentResourceState{
			OrgID:              orgID,
			InstallID:          installID,
			InstallComponentID: installComponentID,
		}).
		Where(app.LatestReportOnlySQL(), app.LatestReportOnlyArgs(orgID, installID)...).
		Where("health != ?", string(app.InstallComponentHealthStatusHealthy)).
		Order("kind, namespace, name").
		Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("unable to query non-healthy resources: %w", err)
	}
	return resources, nil
}

func (s *service) firstHealthObservedAt(ctx context.Context, orgID, installID string) (time.Time, error) {
	rows := make([]app.InstallComponentHealthTransition, 0, 1)
	if err := s.chDB.WithContext(ctx).
		Select("observed_at").
		Where(app.InstallComponentHealthTransition{OrgID: orgID, InstallID: installID}).
		Order("observed_at ASC").
		Limit(1).
		Find(&rows).Error; err != nil {
		return time.Time{}, fmt.Errorf("unable to query first health transition: %w", err)
	}
	if len(rows) == 0 {
		return time.Time{}, nil
	}
	return rows[0].ObservedAt, nil
}
