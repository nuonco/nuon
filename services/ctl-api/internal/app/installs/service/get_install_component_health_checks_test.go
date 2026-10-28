package service

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *LatestReportTestSuite) declareHealth(installComponentID string, probes app.ComponentHealthProbes, required app.ComponentHealthRequiredChecks) {
	ctx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	ctx = cctx.SetOrgIDContext(ctx, s.orgID)

	var ic app.InstallComponent
	require.NoError(s.T(), s.service.DB.WithContext(ctx).First(&ic, "id = ?", installComponentID).Error)

	var appConfigID string
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&app.Install{}).
		Select("app_config_id").
		Where("id = ?", s.installID).
		Scan(&appConfigID).Error)
	require.NotEmpty(s.T(), appConfigID)

	ccc := s.service.Seeder.CreateHelmComponentConfigConnection(ctx, s.T(), ic.ComponentID, appConfigID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where("id = ?", ccc.ID).
		Updates(map[string]any{
			"health_probes":          probes,
			"health_required_checks": required,
		}).Error)
}

func (s *LatestReportTestSuite) fetchHealthChecks(installComponentID string) map[string]app.InstallComponentResourceState {
	var ic app.InstallComponent
	require.NoError(s.T(), s.service.DB.First(&ic, "id = ?", installComponentID).Error)

	got, err := s.service.InstallsService.getInstallComponentHealthChecks(
		context.Background(), s.orgID, s.installID, ic.ComponentID)
	require.NoError(s.T(), err)

	byName := make(map[string]app.InstallComponentResourceState, len(got))
	for _, c := range got {
		byName[c.Name] = c
	}
	return byName
}

func (s *LatestReportTestSuite) TestHealthChecksReturnReportedProbe() {
	s.declareHealth(s.componentA,
		app.ComponentHealthProbes{{Type: "http", Name: "api-healthz", URL: "http://api.default.svc/healthz"}},
		app.ComponentHealthRequiredChecks{"vendor-migration"},
	)
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", resourceProviderProbe, "HTTPProbe", "api-healthz", "healthy", s.latest, 0),
	})

	checks := s.fetchHealthChecks(s.componentA)

	probe, ok := checks["api-healthz"]
	require.True(s.T(), ok)
	assert.Equal(s.T(), resourceProviderProbe, probe.Provider)
	assert.Equal(s.T(), "HTTPProbe", probe.Kind)
	assert.Equal(s.T(), "healthy", probe.Health)
	assert.Empty(s.T(), probe.Message)

	required, ok := checks["vendor-migration"]
	require.True(s.T(), ok)
	assert.Equal(s.T(), customCheckProvider, required.Provider)
	assert.Equal(s.T(), "unknown", required.Health)
	assert.Equal(s.T(), "declared but not reported yet", required.Message)
}

func (s *LatestReportTestSuite) TestHealthChecksUnreportedProbeIsUnknown() {
	s.declareHealth(s.componentA,
		app.ComponentHealthProbes{{Type: "tcp", Name: "db-port", URL: "tcp://db.default.svc:5432"}},
		nil,
	)

	checks := s.fetchHealthChecks(s.componentA)

	probe, ok := checks["db-port"]
	require.True(s.T(), ok)
	assert.Equal(s.T(), resourceProviderProbe, probe.Provider)
	assert.Equal(s.T(), "TCPProbe", probe.Kind)
	assert.Equal(s.T(), "unknown", probe.Health)
	assert.Equal(s.T(), "declared but not reported yet", probe.Message)
}

func (s *LatestReportTestSuite) TestHealthChecksExcludeClusterResources() {
	s.declareHealth(s.componentA, nil, nil)
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "api", "healthy", s.latest, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", app.InstallComponentResourceProviderCustom, "CustomCheck", "pushed", "healthy", s.latest, 1800),
	})

	checks := s.fetchHealthChecks(s.componentA)

	assert.NotContains(s.T(), checks, "api")
	assert.Contains(s.T(), checks, "pushed")
}
