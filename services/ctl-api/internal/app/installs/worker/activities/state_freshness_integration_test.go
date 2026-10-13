package activities_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (s *InstallStateIntegrationSuite) backendPID(tx *gorm.DB) int {
	s.T().Helper()
	var pid int
	require.NoError(s.T(), tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error)
	return pid
}

func (s *InstallStateIntegrationSuite) blockedBy(blocker int, done <-chan struct{}) bool {
	s.T().Helper()
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for {
		var waiters int64
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
			Raw("SELECT count(*) FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid))", blocker).Scan(&waiters).Error)
		if waiters > 0 {
			return true
		}
		select {
		case <-done:
			return false
		case <-deadline.C:
			s.FailNow("the concurrent operation neither finished nor blocked")
		case <-poll.C:
		}
	}
}

func (s *InstallStateIntegrationSuite) await(done <-chan struct{}) {
	s.T().Helper()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		s.FailNow("timed out waiting for the concurrent operation")
	}
}

func (s *InstallStateIntegrationSuite) publicationLocksFree() error {
	return s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
		var head app.InstallStateHead
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			Where(app.InstallStateHead{InstallID: s.install.ID}).First(&head).Error; err != nil {
			return fmt.Errorf("install state head locked during upload: %w", err)
		}
		var install app.Install
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			Select("id").Where(app.Install{ID: s.install.ID}).First(&install).Error; err != nil {
			return fmt.Errorf("install locked during upload: %w", err)
		}
		return nil
	})
}

func (s *InstallStateIntegrationSuite) TestStaleBaselinePublicationIsRejected() {
	clean := s.baseline()
	s.updateInput()
	_, err := s.publish(s.ctx, clean, pkgstate.PartialInputs)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)

	marked := s.baseline()
	require.NotNil(s.T(), marked.State.StaleAt)
	require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialInputs}, marked.Head.StalePartials)
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{"region": generics.ToPtr("ap-south-1")})
	_, err = s.publish(s.ctx, marked, pkgstate.PartialInputs)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "a second mutation to a dirty partial must fence its baseline: %v", err)
	require.Equal(s.T(), marked.Head.Revision+1, s.head().Revision)

	require.NoError(s.T(), s.renameApp("after"))
	pending := s.baseline()
	require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialApp, pkgstate.PartialInputs}, pending.Head.StalePartials)
	pending.State.Inputs = &state.InputsState{Inputs: map[string]string{"region": "ap-south-1"}}
	_, err = s.publish(s.ctx, pending, pkgstate.PartialInputs)
	require.NoError(s.T(), err)
	head := s.head()
	require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialApp}, head.StalePartials)
	require.True(s.T(), pending.Head.StaleAt.Time.Equal(head.StaleAt.Time), "the unrefreshed marker must keep its original instant")

	current := s.fresh()
	require.Equal(s.T(), "after", current.App.Name)
	require.Equal(s.T(), "ap-south-1", current.Inputs.Inputs["region"])
	require.Empty(s.T(), s.head().StalePartials)
}

func (s *InstallStateIntegrationSuite) TestConcurrentPublishersFromOneBaseline() {
	first, second := s.baseline(), s.baseline()
	first.State.Name, second.State.Name = "first-candidate", "second-candidate"
	start := make(chan struct{})
	type result struct {
		row *app.InstallState
		err error
	}
	results := make(chan result, 2)
	for _, candidate := range []*helpers.InstallStateBaseline{first, second} {
		go func() {
			<-start
			row, err := s.publish(s.ctx, candidate)
			results <- result{row, err}
		}()
	}
	close(start)
	var winner *app.InstallState
	for range 2 {
		var r result
		select {
		case r = <-results:
		case <-time.After(30 * time.Second):
			s.FailNow("timed out waiting for publishers")
		}
		if r.err != nil {
			require.True(s.T(), helpers.IsInstallStateConflict(r.err), "%v", r.err)
			continue
		}
		require.Nil(s.T(), winner, "two candidates from one baseline must not both publish")
		winner = r.row
	}
	require.NotNil(s.T(), winner)
	head := s.head()
	require.Equal(s.T(), first.Head.Revision+1, head.Revision)
	require.Equal(s.T(), winner.ID, head.CurrentStateID)
	require.Equal(s.T(), winner.State.Name, s.fresh().Name)
}

func (s *InstallStateIntegrationSuite) TestSourceChangeDuringUploadFencesPublication() {
	before := s.snapshotCount()
	_, err := s.publish(s.withUploadHook(func() error { return s.renameApp("changed-during-upload") }), s.baseline())
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	require.Equal(s.T(), before, s.snapshotCount())
	require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialApp}, s.head().StalePartials)
	require.Equal(s.T(), "changed-during-upload", s.fresh().App.Name)
}

func (s *InstallStateIntegrationSuite) TestRefreshRetriesWithoutHoldingLocksDuringUpload() {
	require.NoError(s.T(), s.renameApp("first-update"))
	uploads := 0
	ctx := s.withUploadHook(func() error {
		uploads++
		if err := s.publicationLocksFree(); err != nil {
			return err
		}
		if uploads == 1 {
			return s.renameApp("second-update")
		}
		return nil
	})
	current, err := s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, false)
	require.NoError(s.T(), err)
	require.Nil(s.T(), current.StaleAt)
	require.Equal(s.T(), "second-update", current.App.Name)
	require.Equal(s.T(), 2, uploads)
	require.Equal(s.T(), "second-update", s.fresh().App.Name)
}

func (s *InstallStateIntegrationSuite) TestContinuousSourceChangesFailInsteadOfServingStale() {
	require.NoError(s.T(), s.renameApp("initial-update"))
	changes := 0
	ctx := s.withUploadHook(func() error {
		changes++
		return s.renameApp(fmt.Sprintf("change-%d", changes))
	})
	current, err := s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, false)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	require.Nil(s.T(), current)
	_, err = s.activities.GetInstallState(ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.Error(s.T(), err)
	require.False(s.T(), s.head().StaleAt.Empty())
	require.Equal(s.T(), fmt.Sprintf("change-%d", changes), s.fresh().App.Name)
}

func (s *InstallStateIntegrationSuite) TestSourceRollbackDiscardsInvalidation() {
	before := s.head()
	rollback := errors.New("abort source update")
	err := s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&app.App{ID: s.install.AppID}).Update("name", "uncommitted").Error; err != nil {
			return err
		}
		var inside app.InstallStateHead
		if err := tx.Where(app.InstallStateHead{InstallID: s.install.ID}).First(&inside).Error; err != nil {
			return err
		}
		require.Equal(s.T(), before.Revision+1, inside.Revision)
		require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialApp}, inside.StalePartials)
		visible := s.baseline()
		require.Equal(s.T(), before, *visible.Head)
		require.Nil(s.T(), visible.State.StaleAt)
		return rollback
	})
	require.ErrorIs(s.T(), err, rollback)
	require.Equal(s.T(), before, s.head())
	require.NotEqual(s.T(), "uncommitted", s.fresh().App.Name)
	_, err = s.publish(s.ctx, s.baseline())
	require.NoError(s.T(), err)
}

func (s *InstallStateIntegrationSuite) TestConcurrentInvalidationsUnionWithinInstallScope() {
	var application app.App
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.App{ID: s.install.AppID}).First(&application).Error)
	sibling := s.deps.Seeder.CreateInstall(s.ctx, s.T(), &application)
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), otherApp.ID)
	unrelated := s.deps.Seeder.CreateInstall(s.ctx, s.T(), otherApp)
	heads := map[string]app.InstallStateHead{}
	for _, installID := range []string{s.install.ID, sibling.ID, unrelated.ID} {
		_, err := s.deps.Helpers.GetInstallState(s.ctx, installID, false, false)
		require.NoError(s.T(), err)
		var head app.InstallStateHead
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: installID}).First(&head).Error)
		heads[installID] = head
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, commit := range []func() error{
		func() error { return s.renameApp("shared-app-update") },
		func() error {
			return s.deps.DB.WithContext(s.ctx).Create(&app.InstallInputs{
				InstallID: s.install.ID, AppInputConfigID: s.inputCfg.ID,
				Values: pgtype.Hstore{"region": generics.ToPtr("eu-west-1")},
			}).Error
		},
	} {
		go func() {
			<-start
			errs <- commit()
		}()
	}
	close(start)
	for range 2 {
		select {
		case err := <-errs:
			require.NoError(s.T(), err)
		case <-time.After(30 * time.Second):
			s.FailNow("timed out waiting for source commits")
		}
	}

	head := s.head()
	require.Equal(s.T(), heads[s.install.ID].Revision+2, head.Revision)
	require.Equal(s.T(), []pkgstate.PartialName{pkgstate.PartialApp, pkgstate.PartialInputs}, head.StalePartials)
	for _, installID := range []string{s.install.ID, sibling.ID} {
		current, err := s.deps.Helpers.GetInstallState(s.ctx, installID, false, false)
		require.NoError(s.T(), err)
		require.Nil(s.T(), current.StaleAt)
		require.Equal(s.T(), "shared-app-update", current.App.Name)
	}
	require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"])
	var untouched app.InstallStateHead
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: unrelated.ID}).First(&untouched).Error)
	require.Equal(s.T(), heads[unrelated.ID], untouched)
}

func (s *InstallStateIntegrationSuite) TestArchiveRetainsCurrentSnapshot() {
	current := s.currentRow()
	newer := make([]string, 0, 55)
	for range 55 {
		row := &app.InstallState{
			InstallID: s.install.ID, Revision: current.Revision,
			State: &state.State{ID: s.install.ID, Name: "historical-snapshot"},
		}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(row).Error)
		newer = append(newer, row.ID)
	}
	require.NoError(s.T(), s.activities.ArchiveState(s.ctx, &installactivities.ArchiveStateRequest{InstallID: s.install.ID}))

	var retained app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{ID: current.ID}).First(&retained).Error)
	require.False(s.T(), retained.Archived)
	require.NotNil(s.T(), retained.State)
	var archived int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where("id IN ?", newer).Where(app.InstallState{Archived: true}).Count(&archived).Error)
	require.Equal(s.T(), int64(5), archived)

	s.deps.Cfg.BlobReadEnabled = false
	snapshot := s.fresh()
	require.Equal(s.T(), s.install.Name, snapshot.Name)
	require.Equal(s.T(), "us-east-1", snapshot.Inputs.Inputs["region"])
}

func (s *InstallStateIntegrationSuite) TestComponentDeletionFencesPublication() {
	connection := s.componentConnection()
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
	for _, hard := range []bool{false, true} {
		require.Contains(s.T(), s.fresh().Components, connection.Component.Name)
		stale := s.baseline()
		db := s.deps.DB.WithContext(s.ctx)
		if hard {
			db = db.Unscoped()
		}
		require.NoError(s.T(), db.Delete(installed).Error)
		_, err := s.publish(s.ctx, stale, pkgstate.PartialComponents)
		require.True(s.T(), helpers.IsInstallStateConflict(err), "hard=%t: %v", hard, err)
		require.NotContains(s.T(), s.fresh().Components, connection.Component.Name)
		if !hard {
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Unscoped().Model(installed).Update("deleted_at", 0).Error)
			installed.DeletedAt = 0
		}
	}
}

func (s *InstallStateIntegrationSuite) TestInstallSoftDeletionFencesPublication() {
	stale := s.baseline()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.Install{ID: s.install.ID}).Error)
	require.Equal(s.T(), stale.Head.Revision+1, s.head().Revision)
	before := s.snapshotCount()
	_, err := s.publish(s.ctx, stale)
	require.True(s.T(), helpers.IsInstallStateConflict(err) || errors.Is(err, gorm.ErrRecordNotFound), "%v", err)
	require.Equal(s.T(), before, s.snapshotCount())
	require.Equal(s.T(), pkgstate.AllPartials, s.head().StalePartials)
	_, err = s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.Error(s.T(), err, "a deleted install must not be served from its last snapshot")
}

func (s *InstallStateIntegrationSuite) TestInstallHardDeletionFencesConcurrentPublication() {
	baseline := s.baseline()
	tx := s.deps.DB.WithContext(s.ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), tx.Unscoped().Where(app.InstallStackOutputs{InstallStackID: s.install.InstallStack.ID}).Delete(&app.InstallStackOutputs{}).Error)
	require.NoError(s.T(), tx.Unscoped().Scopes(scopes.WithDisableViews).Where(app.Install{ID: s.install.ID}).Delete(&app.Install{}).Error)

	pid := s.backendPID(tx)
	done := make(chan struct{})
	var publishErr error
	go func() {
		defer close(done)
		_, publishErr = s.publish(s.ctx, baseline)
	}()
	require.True(s.T(), s.blockedBy(pid, done), "publication must wait for the install deletion")
	require.NoError(s.T(), tx.Commit().Error)
	s.await(done)
	require.ErrorIs(s.T(), publishErr, gorm.ErrRecordNotFound)
	require.Zero(s.T(), s.snapshotCount())
	var heads int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallStateHead{}).
		Where(app.InstallStateHead{InstallID: s.install.ID}).Count(&heads).Error)
	require.Zero(s.T(), heads)
}

func (s *InstallStateIntegrationSuite) TestPublicationAndHardDeletionDoNotDeadlock() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Unscoped().
		Where(app.InstallStackOutputs{InstallStackID: s.install.InstallStack.ID}).Delete(&app.InstallStackOutputs{}).Error)
	s.fresh()
	baseline := s.baseline()
	blocker := s.deps.DB.WithContext(s.ctx).Begin()
	require.NoError(s.T(), blocker.Error)
	defer blocker.Rollback()
	var head app.InstallStateHead
	require.NoError(s.T(), blocker.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).
		Where(app.InstallStateHead{InstallID: s.install.ID}).First(&head).Error)
	blockerPID := s.backendPID(blocker)

	published := make(chan struct{})
	var publishErr error
	go func() {
		defer close(published)
		_, publishErr = s.publish(s.ctx, baseline)
	}()
	require.True(s.T(), s.blockedBy(blockerPID, published))
	var publisherPID int
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Raw("SELECT pid FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid))", blockerPID).Scan(&publisherPID).Error)
	require.NotZero(s.T(), publisherPID)

	deleted := make(chan struct{})
	var deleteErr error
	go func() {
		defer close(deleted)
		deleteErr = s.deps.DB.WithContext(s.ctx).Unscoped().Scopes(scopes.WithDisableViews).
			Where(app.Install{ID: s.install.ID}).Delete(&app.Install{}).Error
	}()
	require.True(s.T(), s.blockedBy(publisherPID, deleted), "deletion must wait for the publishing transaction's install lock")
	require.NoError(s.T(), blocker.Commit().Error)
	s.await(published)
	s.await(deleted)
	require.NoError(s.T(), publishErr)
	require.NoError(s.T(), deleteErr)
	require.Zero(s.T(), s.snapshotCount())
}
