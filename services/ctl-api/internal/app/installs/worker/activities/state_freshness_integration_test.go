package activities_test

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (s *InstallStateIntegrationSuite) TestInputRefreshDoesNotEraseWholeStateInvalidation() {
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.App = &state.AppState{ID: s.install.AppID, Name: "before"}
	})
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "after").Error)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStateStale(s.ctx, s.install.ID))
	stale, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), stale.App)
	require.True(s.T(), stale.App.Name == "after" || stale.StaleAt != nil)
	s.updateInput()
	for read := 0; read < 2; read++ {
		snapshot, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"])
		require.NotNil(s.T(), snapshot.App)
		require.True(s.T(), snapshot.App.Name == "after" || snapshot.StaleAt != nil,
			"refreshing inputs must not declare an old app name fresh on read %d", read)
	}
}

func (s *InstallStateIntegrationSuite) TestLatePartialSavePreservesWholeStateInvalidation() {
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.App = &state.AppState{ID: s.install.AppID, Name: "before"}
	})
	older, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	started := time.Now().UTC().Add(-time.Second)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "after").Error)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStateStale(s.ctx, s.install.ID))
	_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
		State: older, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy:       app.InstallStateGenerateSourceStateManager,
		RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs}, RefreshStartedAt: started,
	})
	if err != nil {
		s.T().Logf("late generation rejected: %v", err)
	}
	latest, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), latest.App)
	require.True(s.T(), latest.App.Name == "after" || latest.StaleAt != nil,
		"an inputs-only generation predating full invalidation must not declare an old app name fresh")
}

func (s *InstallStateIntegrationSuite) TestNamedInvalidationPublicationBoundary() {
	for _, tc := range []struct {
		name       string
		startDelta time.Duration
		wantStale  bool
	}{
		{name: "marker_before_refresh", startDelta: time.Second},
		{name: "marker_equal_to_refresh"},
		{name: "marker_after_refresh", startDelta: -time.Second, wantStale: true},
	} {
		s.Run(tc.name, func() {
			baseline, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			require.NoError(s.T(), err)
			require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs))
			var marked app.InstallState
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
				Where(app.InstallState{InstallID: s.install.ID}).Order("created_at DESC").First(&marked).Error)
			_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
				State: baseline, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				GeneratedBy:       app.InstallStateGenerateSourceStateManager,
				RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs},
				RefreshStartedAt:  marked.StaleAt.Time.Add(tc.startDelta),
			})
			require.NoError(s.T(), err)
			var latest app.InstallState
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
				Where(app.InstallState{InstallID: s.install.ID}).Order("created_at DESC").First(&latest).Error)
			if tc.wantStale {
				require.False(s.T(), latest.StaleAt.Empty())
				require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialInputs}, latest.StalePartials)
			} else {
				require.True(s.T(), latest.StaleAt.Empty())
				require.Empty(s.T(), latest.StalePartials)
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) TestPartialSavePreservesUnrefreshedInvalidation() {
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.App = &state.AppState{ID: s.install.AppID, Name: "before"}
	})
	baseline, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "after").Error)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialApp, pkgstate.PartialInputs))
	var marked app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallState{InstallID: s.install.ID}).Order("created_at DESC").First(&marked).Error)
	_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
		State: baseline, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy:       app.InstallStateGenerateSourceStateManager,
		RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs}, RefreshStartedAt: marked.StaleAt.Time.Add(time.Second),
	})
	require.NoError(s.T(), err)
	latest, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "after", latest.App.Name, "unrefreshed app invalidation must survive until the normal reader refreshes it")
	require.Nil(s.T(), latest.StaleAt)
}

func (s *InstallStateIntegrationSuite) TestOverlappingGenerationsPreserveNewerComponentOutputs() {
	for _, tc := range []struct {
		name         string
		olderPartial pkgstate.PartialName
	}{
		{name: "different_partials", olderPartial: pkgstate.PartialInputs},
		{name: "same_partial", olderPartial: pkgstate.PartialComponents},
	} {
		s.Run(tc.name, func() {
			s.saveBaseline(func(snapshot *state.State) {
				snapshot.Components = map[string]any{"api": map[string]any{"outputs": map[string]any{"endpoint": "old.example.com"}}}
			})
			older, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			require.NoError(s.T(), err)
			olderStarted := time.Now().UTC().Add(-time.Second)
			newer, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			require.NoError(s.T(), err)
			newer.Components = map[string]any{"api": map[string]any{"outputs": map[string]any{"endpoint": "new.example.com"}}}
			newerStarted := olderStarted.Add(time.Second)
			_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
				State: newer, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				GeneratedBy:       app.InstallStateGenerateSourceStateManager,
				RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialComponents}, RefreshStartedAt: newerStarted,
			})
			require.NoError(s.T(), err)
			current, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			require.NoError(s.T(), err)
			require.Equal(s.T(), newer.Components, current.Components)
			_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
				State: older, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				GeneratedBy:       app.InstallStateGenerateSourceStateManager,
				RefreshedPartials: []pkgstate.PartialName{tc.olderPartial}, RefreshStartedAt: olderStarted,
			})
			if err != nil {
				s.T().Logf("late generation rejected: %v", err)
			}
			for read := 0; read < 2; read++ {
				latest, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
				require.NoError(s.T(), err)
				require.Equal(s.T(), newer.Components, latest.Components,
					"an older %s generation publishing last must not restore old component outputs", tc.olderPartial)
			}
		})
	}
}
