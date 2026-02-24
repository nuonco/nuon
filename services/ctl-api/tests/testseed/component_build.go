package testseed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CreateComponentBuild persists a ComponentBuild linked to the given ComponentConfigConnection.
func (s *Seeder) CreateComponentBuild(ctx context.Context, t *testing.T, configConnectionID string) *app.ComponentBuild {
	bld := &app.ComponentBuild{
		ComponentConfigConnectionID: configConnectionID,
		Status:                      app.ComponentBuildStatusPlanning,
		StatusDescription:           "queued and waiting for runner to pick up",
	}
	res := s.db.WithContext(ctx).Create(bld)
	require.NoError(t, res.Error)
	return bld
}
