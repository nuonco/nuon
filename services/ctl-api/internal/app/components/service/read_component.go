package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

func (s *service) readComponent(ctx context.Context, targetName string) (*app.Component, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := access.Target(targetName)
	if err != nil {
		return nil, err
	}
	component, err := s.helpers.GetComponentInTx(ctx, access.DB(), target.ID)
	if err != nil {
		return nil, err
	}
	for _, config := range component.ComponentConfigs {
		if config.OrgID != access.OrgID() || config.ComponentID != component.ID {
			return nil, gorm.ErrRecordNotFound
		}
	}
	dependencies := make([]*app.Component, 0, len(component.Dependencies))
	dependencyIDs := make([]string, 0, len(component.Dependencies))
	for _, dependency := range component.Dependencies {
		if dependency.OrgID == access.OrgID() && dependency.AppID == component.AppID {
			dependencies = append(dependencies, dependency)
			dependencyIDs = append(dependencyIDs, dependency.ID)
		}
	}
	component.Dependencies = dependencies
	component.DependencyIDs = dependencyIDs
	component.Org = app.Org{}
	component.App = nil
	return component, nil
}
