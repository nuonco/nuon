package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) GetComponentLatestBuilds(ctx context.Context, cmpIDs ...string) ([]app.ComponentBuild, error) {
	if len(cmpIDs) == 0 {
		return []app.ComponentBuild{}, nil
	}

	builds := make([]app.ComponentBuild, 0, len(cmpIDs))

	// TODO: we shoould be able to do this in a single query, but for no just avoiding the previous large result from
	// the previous implementation
	for _, cmpID := range cmpIDs {
		build, err := s.getComponentLatestBuild(ctx, cmpID)
		if err != nil {
			// Skip components that don't have builds instead of failing entirely
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		builds = append(builds, *build)
	}

	return builds, nil
}

func (s *Helpers) getComponentLatestBuild(ctx context.Context, cmpID string) (*app.ComponentBuild, error) {
	var build app.ComponentBuild

	res := s.db.WithContext(ctx).
		Joins("JOIN component_config_connections ON component_builds.component_config_connection_id = component_config_connections.id").
		Where("component_config_connections.component_id = ?", cmpID).
		Order("component_config_connections.created_at DESC").
		Preload("ComponentConfigConnection").
		Preload("VCSConnectionCommit").
		First(&build)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get latest component build: %w", res.Error)
	}

	return &build, nil
}
