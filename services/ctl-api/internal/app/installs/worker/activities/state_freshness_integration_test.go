package activities_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
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
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs))
	_, err = s.saveState(baseline)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	marked, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), marked.StaleAt)
	require.Equal(s.T(), []string{"inputs"}, marked.StalePartials)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs))
	_, err = s.saveState(marked)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "a second mutation to an already-dirty section must invalidate its baseline: %v", err)
	current, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Nil(s.T(), current.StaleAt)
}

func (s *InstallStateIntegrationSuite) TestPartialSavePreservesUnrefreshedInvalidation() {
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.App = &state.AppState{ID: s.install.AppID, Name: "before"}
	})
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "after").Error)
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialApp, pkgstate.PartialInputs))
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	_, err = s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
		State: baseline, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy:       app.InstallStateGenerateSourceStateManager,
		RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs},
	})
	require.NoError(s.T(), err)
	pending, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), []string{"app"}, pending.StalePartials)
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

type publicationBlobService struct {
	blobstore.Service
	afterUpload func() error
}

func (b publicationBlobService) UploadStream(ctx context.Context, key string, data io.Reader) (string, error) {
	checksum, err := b.Service.UploadStream(ctx, key, data)
	if err == nil {
		err = b.afterUpload()
	}
	return checksum, err
}

func (s *InstallStateIntegrationSuite) TestGenerationRejectsSourceChangeDuringUpload() {
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	ctx := blobstore.WithBlobService(s.ctx, publicationBlobService{
		Service: s.deps.Blobs,
		afterUpload: func() error {
			return s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "changed-during-upload").Error
		},
	})
	_, err = s.activities.SaveState(ctx, &installactivities.SaveStateRequest{
		State: baseline, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy: app.InstallStateGenerateSourceStateManager,
	})
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	var publishedCount, stagedCount int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where(app.InstallState{InstallID: s.install.ID}).Count(&publishedCount).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Scopes(scopes.WithDisableViews).Model(&app.InstallState{}).
		Where(app.InstallState{InstallID: s.install.ID}).Count(&stagedCount).Error)
	require.Equal(s.T(), int64(1), publishedCount)
	require.Equal(s.T(), int64(2), stagedCount)
	current, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "changed-during-upload", current.App.Name)
	require.Nil(s.T(), current.StaleAt)
}

func (s *InstallStateIntegrationSuite) TestRefreshRebuildsAfterSourceChangeDuringUpload() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "first-update").Error)
	var once sync.Once
	ctx := blobstore.WithBlobService(s.ctx, publicationBlobService{
		Service: s.deps.Blobs,
		afterUpload: func() error {
			var err error
			once.Do(func() {
				err = s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "second-update").Error
			})
			return err
		},
	})
	current, err := s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "second-update", current.App.Name)
	require.Nil(s.T(), current.StaleAt)
	current, err = s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "second-update", current.App.Name)
}

func (s *InstallStateIntegrationSuite) TestSourceMutationAndInvalidationRollbackTogether() {
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	rollback := errors.New("abort source update")
	err = s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&app.App{ID: s.install.AppID}).Update("name", "uncommitted").Error; err != nil {
			return err
		}
		var head app.InstallStateHead
		if err := tx.Where(app.InstallStateHead{InstallID: s.install.ID}).First(&head).Error; err != nil {
			return err
		}
		require.Greater(s.T(), head.Revision, *baseline.Revision)
		require.Contains(s.T(), head.StalePartials, pkgstate.PartialApp)
		visible, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
		if err != nil {
			return err
		}
		require.Equal(s.T(), baseline.Revision, visible.Revision)
		require.Nil(s.T(), visible.StaleAt)
		return rollback
	})
	require.ErrorIs(s.T(), err, rollback)
	current, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), baseline.Revision, current.Revision)
	require.Nil(s.T(), current.StaleAt)
	var application app.App
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.App{ID: s.install.AppID}).First(&application).Error)
	require.NotEqual(s.T(), "uncommitted", application.Name)
	_, err = s.saveState(baseline)
	require.NoError(s.T(), err)
}

func (s *InstallStateIntegrationSuite) TestOnlyOneConcurrentGeneratorPublishes() {
	first, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	second, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	first.Name, second.Name = "first-candidate", "second-candidate"
	start := make(chan struct{})
	results := make(chan struct {
		row *app.InstallState
		err error
	}, 2)
	for _, candidate := range []*state.State{first, second} {
		go func(snapshot *state.State) {
			<-start
			row, err := s.saveState(snapshot)
			results <- struct {
				row *app.InstallState
				err error
			}{row, err}
		}(candidate)
	}
	close(start)
	var winner *app.InstallState
	for attempt := 0; attempt < 2; attempt++ {
		result := <-results
		if result.err == nil {
			require.Nil(s.T(), winner, "two candidates from one baseline must not both publish")
			winner = result.row
		} else {
			require.True(s.T(), helpers.IsInstallStateConflict(result.err), "%v", result.err)
		}
	}
	require.NotNil(s.T(), winner)
	current, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), winner.State.Name, current.Name)
	require.Equal(s.T(), *first.Revision+1, *current.Revision)
}

func (s *InstallStateIntegrationSuite) TestLegacySnapshotBootstrapsWithoutBlobRewrite() {
	legacy := &state.State{ID: s.install.ID, Name: "legacy-install", Inputs: &state.InputsState{Inputs: map[string]string{"region": "us-east-1"}}}
	helpers.MapLegacyFields(legacy)
	encoded, err := json.Marshal(legacy)
	require.NoError(s.T(), err)
	blob := &blobstore.Blob{}
	blob.Set(string(encoded))
	row := &app.InstallState{InstallID: s.install.ID, State: legacy, StateBlob: blob, GeneratedBy: app.InstallStateGenerateSourceLegacy}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(row).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: s.install.ID}).Delete(&app.InstallStateHead{}).Error)
	for _, blobRead := range []bool{true, false} {
		s.deps.Cfg.BlobReadEnabled = blobRead
		current, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
		require.NoError(s.T(), err)
		require.Equal(s.T(), "legacy-install", current.Name)
		require.Equal(s.T(), "us-east-1", current.Inputs.Inputs["region"])
		require.NotNil(s.T(), current.Revision)
		require.Nil(s.T(), current.StaleAt)
	}
	var stored app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{ID: row.ID}).First(&stored).Error)
	require.Equal(s.T(), row.StateBlob.BlobID(), stored.StateBlob.BlobID())
	data, err := json.Marshal(legacy)
	require.NoError(s.T(), err)
	require.NotContains(s.T(), string(data), "state_revision")
	require.NotContains(s.T(), string(data), "stale_partials")
	_, err = s.saveState(legacy)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "old activity payloads without a baseline must fail safely: %v", err)
}

func (s *InstallStateIntegrationSuite) TestBulkInvalidationAndConcurrentSectionsRemainInstallScoped() {
	var application app.App
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.App{ID: s.install.AppID}).First(&application).Error)
	second := s.deps.Seeder.CreateInstall(s.ctx, s.T(), &application)
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), otherApp.ID)
	unrelated := s.deps.Seeder.CreateInstall(s.ctx, s.T(), otherApp)
	baseline := make(map[string]*state.State)
	for _, installID := range []string{s.install.ID, second.ID, unrelated.ID} {
		current, err := s.deps.Helpers.GetInstallState(s.ctx, installID, false, false)
		require.NoError(s.T(), err)
		baseline[installID] = current
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, partial := range []pkgstate.PartialName{pkgstate.PartialInputs, pkgstate.PartialComponents} {
		go func(partial pkgstate.PartialName) {
			<-start
			results <- s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, partial)
		}(partial)
	}
	close(start)
	require.NoError(s.T(), <-results)
	require.NoError(s.T(), <-results)
	current, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), []string{"inputs", "components"}, current.StalePartials)
	require.Equal(s.T(), *baseline[s.install.ID].Revision+2, *current.Revision)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&application).Update("name", "shared-app-update").Error)
	for _, installID := range []string{s.install.ID, second.ID} {
		pending, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, installID)
		require.NoError(s.T(), err)
		require.Contains(s.T(), pending.StalePartials, "app")
		current, err := s.deps.Helpers.GetInstallState(s.ctx, installID, false, false)
		require.NoError(s.T(), err)
		require.Equal(s.T(), "shared-app-update", current.App.Name)
		require.Nil(s.T(), current.StaleAt)
	}
	current, err = s.deps.Helpers.GetInstallStateBaseline(s.ctx, unrelated.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), baseline[unrelated.ID].Revision, current.Revision)
	require.Nil(s.T(), current.StaleAt)
}

func (s *InstallStateIntegrationSuite) TestSourceDeletionRemovesComponentFromFreshState() {
	var connection app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Preload("Component").
		Where(app.ComponentConfigConnection{AppConfigID: s.install.AppConfigID, Type: app.ComponentTypeTerraformModule}).First(&connection).Error)
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
	for _, hardDelete := range []bool{false, true} {
		current, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
		require.NoError(s.T(), err)
		require.Contains(s.T(), current.Components, connection.Component.Name)
		db := s.deps.DB.WithContext(s.ctx)
		if hardDelete {
			db = db.Unscoped()
		}
		require.NoError(s.T(), db.Delete(installed).Error)
		pending, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
		require.NoError(s.T(), err)
		require.Contains(s.T(), pending.StalePartials, "components")
		current, err = s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
		require.NoError(s.T(), err)
		require.NotContains(s.T(), current.Components, connection.Component.Name)
		require.Nil(s.T(), current.StaleAt)
		if !hardDelete {
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Unscoped().Model(installed).Update("deleted_at", 0).Error)
			installed.DeletedAt = 0
		}
	}
}

func (s *InstallStateIntegrationSuite) TestRefreshFailsInsteadOfReturningStaleDuringContinuousChanges() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "initial-update").Error)
	changes := 0
	ctx := blobstore.WithBlobService(s.ctx, publicationBlobService{
		Service: s.deps.Blobs,
		afterUpload: func() error {
			changes++
			return s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", fmt.Sprintf("change-%d", changes)).Error
		},
	})
	current, err := s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, false)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	require.Nil(s.T(), current)
	pending, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), pending.StaleAt)
	current, err = s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fmt.Sprintf("change-%d", changes), current.App.Name)
	require.Nil(s.T(), current.StaleAt)
}

func (s *InstallStateIntegrationSuite) TestArchiveRetainsCurrentSnapshotDespiteAbandonedCandidates() {
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	candidateIDs := make([]string, 0, 55)
	for candidate := 0; candidate < 55; candidate++ {
		staged := &app.InstallState{
			InstallID: s.install.ID, Revision: *baseline.Revision,
			State: &state.State{ID: s.install.ID, Name: "unpublished"},
		}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(staged).Error)
		candidateIDs = append(candidateIDs, staged.ID)
	}
	require.NoError(s.T(), s.activities.ArchiveState(s.ctx, &installactivities.ArchiveStateRequest{InstallID: s.install.ID}))
	s.deps.Cfg.BlobReadEnabled = false
	current, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Equal(s.T(), baseline.Name, current.Name)
	require.Equal(s.T(), "us-east-1", current.Inputs.Inputs["region"])
	var visible int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where(app.InstallState{InstallID: s.install.ID}).Count(&visible).Error)
	require.Equal(s.T(), int64(1), visible)
	var candidates []app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Scopes(scopes.WithDisableViews).
		Where(app.InstallState{InstallID: s.install.ID}).Where("id IN ?", candidateIDs).Find(&candidates).Error)
	require.Len(s.T(), candidates, 55)
	for _, candidate := range candidates {
		require.False(s.T(), candidate.Archived)
		require.NotNil(s.T(), candidate.State)
		require.Equal(s.T(), "unpublished", candidate.State.Name)
	}
}
