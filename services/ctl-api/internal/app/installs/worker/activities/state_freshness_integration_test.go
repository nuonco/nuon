package activities_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installservice "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/service"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
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

func (s *InstallStateIntegrationSuite) observeSourceBarrier() <-chan struct{} {
	s.T().Helper()
	attempted := make(chan struct{}, 1)
	require.NoError(s.T(), s.deps.DB.Callback().Row().After("gorm:row").Register("test:source_barrier", func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "pg_try_advisory_xact_lock") {
			select {
			case attempted <- struct{}{}:
			default:
			}
		}
	}))
	s.T().Cleanup(func() { _ = s.deps.DB.Callback().Row().Remove("test:source_barrier") })
	return attempted
}

func (s *InstallStateIntegrationSuite) publicationLocksFree() error {
	return s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
		var available bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(hashtext('install_state_sources'), hashtext(?))", s.install.OrgID).
			Scan(&available).Error; err != nil {
			return err
		}
		if !available {
			return fmt.Errorf("source coordination lock held during upload")
		}
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
	var before, after time.Time
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&before).Error)
	s.updateInput()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Raw("SELECT clock_timestamp()").Scan(&after).Error)
	_, err := s.publish(s.ctx, clean, pkgstate.PartialInputs)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)

	marked := s.baseline()
	require.NotNil(s.T(), marked.State.StaleAt)
	require.True(s.T(), !marked.State.StaleAt.Before(before) && !marked.State.StaleAt.After(after))
	require.True(s.T(), marked.State.StaleAt.Equal(marked.Head.StaleAt.Time))
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

func (s *InstallStateIntegrationSuite) TestRefreshRetriesWithoutHoldingLocksDuringUpload() {
	require.NoError(s.T(), s.renameApp("first-update"))
	before := s.snapshotCount()
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
	require.Equal(s.T(), before+1, s.snapshotCount(), "only the successful candidate may leave a snapshot row")
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
	for name, read := range map[string]func() error{
		"get_activity": func() error {
			_, err := s.activities.GetInstallState(ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			return err
		},
		"regenerate_activity": func() error {
			_, err := s.activities.RegenerateInstallState(ctx, &pkgstate.ExecuteRegenerationRequest{InstallID: s.install.ID})
			return err
		},
	} {
		s.Run(name, func() {
			err := read()
			require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
			failure := temporal.GetDefaultFailureConverter().ErrorToFailure(err).GetApplicationFailureInfo()
			require.NotNil(s.T(), failure)
			require.Equal(s.T(), helpers.InstallStateConflict, failure.Type)
			require.False(s.T(), failure.NonRetryable)
		})
	}
	api, err := installservice.New(installservice.Params{Cfg: s.deps.Cfg, Helpers: s.deps.Helpers})
	require.NoError(s.T(), err)
	response := httptest.NewRecorder()
	httpCtx, engine := gin.CreateTestContext(response)
	engine.ContextWithFallback = true
	httpCtx.Request = httptest.NewRequest(http.MethodGet, "/v1/installs/"+s.install.ID+"/state", nil).WithContext(ctx)
	httpCtx.Params = gin.Params{{Key: "install_id", Value: s.install.ID}}
	api.GetInstallState(httpCtx)
	require.Empty(s.T(), httpCtx.Errors)
	require.Equal(s.T(), http.StatusServiceUnavailable, response.Code)
	require.Equal(s.T(), "1", response.Header().Get("Retry-After"))
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

func (s *InstallStateIntegrationSuite) TestSharedSourceCommitReachesConcurrentNewInstall() {
	for name, tc := range map[string]struct {
		mutate func(*gorm.DB) error
		value  func(*state.State) string
		want   string
	}{
		"app_variable": {
			mutate: func(tx *gorm.DB) error {
				return tx.Create(&app.AppSecret{AppID: s.install.AppID, Name: "external_api_key", Value: "example-value"}).Error
			},
			value: func(snapshot *state.State) string { return snapshot.App.Variables["external_api_key"] },
			want:  "example-value",
		},
		"app_metadata": {
			mutate: func(tx *gorm.DB) error {
				return tx.Model(&app.App{ID: s.install.AppID}).Update("status", app.AppStatusError).Error
			},
			value: func(snapshot *state.State) string { return snapshot.App.Status },
			want:  string(app.AppStatusError),
		},
		"org_metadata": {
			mutate: func(tx *gorm.DB) error {
				return tx.Model(&app.Org{ID: s.install.OrgID}).Update("status", app.OrgStatusError).Error
			},
			value: func(snapshot *state.State) string { return snapshot.Org.Status },
			want:  string(app.OrgStatusError),
		},
		"input_default": {
			mutate: func(tx *gorm.DB) error {
				return tx.Model(&app.AppInput{}).Where(app.AppInput{AppInputConfigID: s.inputCfg.ID, Name: "region"}).
					Update("default", "ap-south-1").Error
			},
			value: func(snapshot *state.State) string { return snapshot.Inputs.Inputs["region"] },
			want:  "ap-south-1",
		},
	} {
		s.Run(name, func() {
			tx := s.deps.DB.WithContext(s.ctx).Begin()
			require.NoError(s.T(), tx.Error)
			defer tx.Rollback()
			require.NoError(s.T(), tc.mutate(tx))

			created := &app.Install{
				Name: "concurrent-" + name, AppID: s.install.AppID, AppConfigID: s.install.AppConfigID,
				AppRunnerConfigID: s.install.AppRunnerConfigID, AppSandboxConfigID: s.install.AppSandboxConfigID,
				AWSAccount: &app.AWSAccount{Region: "us-east-1"},
				InstallInputs: []app.InstallInputs{{
					AppInputConfigID: s.inputCfg.ID, Values: pgtype.Hstore{"region": generics.ToPtr("")},
				}},
			}
			attempted := s.observeSourceBarrier()
			done := make(chan struct{})
			var createErr error
			go func() {
				defer close(done)
				if createErr = s.deps.DB.WithContext(s.ctx).Create(created).Error; createErr == nil {
					_, createErr = s.deps.Helpers.GetInstallState(s.ctx, created.ID, false, true)
				}
			}()
			select {
			case <-done:
			case <-attempted:
			case <-time.After(30 * time.Second):
				s.FailNow("timed out waiting for the concurrent state read")
			}
			require.NoError(s.T(), tx.Commit().Error)
			s.await(done)
			require.NoError(s.T(), createErr)

			for read := 0; read < 2; read++ {
				var head app.InstallStateHead
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: created.ID}).First(&head).Error)
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
					Where(app.InstallState{ID: head.CurrentStateID}).Select("state").Updates(map[string]any{"state": nil}).Error)
				snapshot, err := s.deps.Helpers.GetInstallState(s.ctx, created.ID, false, true)
				require.NoError(s.T(), err)
				require.Nil(s.T(), snapshot.StaleAt)
				s.Equal(tc.want, tc.value(snapshot), "read %d must observe the committed source", read+1)
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) TestSourceWaitPreservesWriterConcurrencyAndCancellation() {
	timeouts := func() int64 {
		var data metricdata.ResourceMetrics
		require.NoError(s.T(), s.reader.Collect(s.ctx, &data))
		var total int64
		for _, scope := range data.ScopeMetrics {
			for _, metric := range scope.Metrics {
				if metric.Name == "nuon.install.state.source_wait.timeouts" {
					for _, point := range metric.Data.(metricdata.Sum[int64]).DataPoints {
						total += point.Value
					}
				}
			}
		}
		return total
	}
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	otherCtx, _ := s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	unrelatedApp := s.deps.Seeder.CreateApp(otherCtx, s.T())
	s.deps.Seeder.CreateAppConfig(otherCtx, s.T(), unrelatedApp.ID)
	unrelated := s.deps.Seeder.CreateInstall(otherCtx, s.T(), unrelatedApp)
	s.updateInput()
	before, previousTimeouts := s.head(), timeouts()
	tx := s.deps.DB.WithContext(s.ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), tx.Create(&app.AppSecret{AppID: s.install.AppID, Name: "pending", Value: "example"}).Error)

	attempted := s.observeSourceBarrier()
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	done := make(chan struct{})
	var snapshot *state.State
	var readErr error
	go func() {
		defer close(done)
		snapshot, readErr = s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, true)
	}()
	select {
	case <-attempted:
	case <-done:
		s.FailNow("the read returned before the pending source transaction finished", "%v", readErr)
	case <-time.After(30 * time.Second):
		s.FailNow("timed out waiting for the source barrier")
	}

	writeCtx, stop := context.WithTimeout(s.ctx, time.Second)
	defer stop()
	require.NoError(s.T(), s.deps.DB.WithContext(writeCtx).Create(&app.AppSecret{AppID: otherApp.ID, Name: "concurrent", Value: "example"}).Error)
	require.NoError(s.T(), tx.Create(&app.AppSecret{AppID: otherApp.ID, Name: "second-statement", Value: "example"}).Error)
	unrelatedCtx, stopUnrelated := context.WithTimeout(otherCtx, time.Second)
	defer stopUnrelated()
	unrelatedState, err := s.deps.Helpers.GetInstallState(unrelatedCtx, unrelated.ID, false, true)
	require.NoError(s.T(), err)
	require.Equal(s.T(), unrelatedApp.ID, unrelatedState.App.ID)
	cancel()
	s.await(done)
	require.ErrorIs(s.T(), readErr, context.Canceled)
	require.Nil(s.T(), snapshot)
	shortCtx, stopShort := context.WithTimeout(s.ctx, 100*time.Millisecond)
	defer stopShort()
	_, err = s.deps.Helpers.GetInstallState(shortCtx, s.install.ID, false, true)
	require.ErrorIs(s.T(), err, context.DeadlineExceeded)
	require.Equal(s.T(), previousTimeouts, timeouts(), "caller cancellation and deadlines are not barrier timeouts")
	snapshot, err = s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, true)
	require.ErrorIs(s.T(), err, context.DeadlineExceeded)
	require.Nil(s.T(), snapshot)
	require.Equal(s.T(), previousTimeouts+1, timeouts())
	require.Equal(s.T(), before, s.head())
	require.NoError(s.T(), tx.Rollback().Error)
	require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"])
	require.NotContains(s.T(), s.fresh().App.Variables, "pending")
	require.Equal(s.T(), previousTimeouts+1, timeouts(), "successful refreshes must not increment timeouts")
}

func (s *InstallStateIntegrationSuite) TestIrrelevantUpdatesDoNotBlockRefresh() {
	otherCtx, _ := s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	otherApp := s.deps.Seeder.CreateApp(otherCtx, s.T())
	s.deps.Seeder.CreateAppConfig(otherCtx, s.T(), otherApp.ID)
	otherInstall := s.deps.Seeder.CreateInstall(otherCtx, s.T(), otherApp)

	for name, mixed := range map[string]bool{"noise_only": false, "mixed_orgs": true} {
		s.Run(name, func() {
			s.updateInput()
			before := s.head()
			tx := s.deps.DB.WithContext(s.ctx).Begin()
			require.NoError(s.T(), tx.Error)
			defer tx.Rollback()
			if mixed {
				require.NoError(s.T(), tx.Exec(`UPDATE apps SET updated_at = now(),
                    name = CASE WHEN id = ? THEN 'changed-app' ELSE name END WHERE id IN (?, ?)`,
					otherApp.ID, s.install.AppID, otherApp.ID).Error)
			} else {
				require.NoError(s.T(), tx.Model(&app.App{ID: s.install.AppID}).Update("updated_at", time.Now()).Error)
			}

			ctx, cancel := context.WithTimeout(s.ctx, time.Second)
			defer cancel()
			current, err := s.deps.Helpers.GetInstallState(ctx, s.install.ID, false, true)
			require.NoError(s.T(), err)
			require.Equal(s.T(), "eu-west-1", current.Inputs.Inputs["region"])
			published := s.head()
			require.Equal(s.T(), before.Revision+1, published.Revision)

			if mixed {
				ctx, cancel := context.WithTimeout(otherCtx, 100*time.Millisecond)
				defer cancel()
				_, err := s.deps.Helpers.GetInstallState(ctx, otherInstall.ID, false, true)
				require.ErrorIs(s.T(), err, context.DeadlineExceeded, "the org with a relevant change must still wait")
			}
			require.NoError(s.T(), tx.Commit().Error)
			require.Equal(s.T(), published, s.head(), "noise must not invalidate the refreshed install")
			if mixed {
				current, err := s.deps.Helpers.GetInstallState(otherCtx, otherInstall.ID, false, true)
				require.NoError(s.T(), err)
				require.Equal(s.T(), "changed-app", current.App.Name)
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) TestSourceWritesRequireReadCommitted() {
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	for name, isolation := range map[string]sql.IsolationLevel{
		"read_committed":  sql.LevelReadCommitted,
		"repeatable_read": sql.LevelRepeatableRead,
		"serializable":    sql.LevelSerializable,
	} {
		s.Run(name, func() {
			before := s.head()
			secret := &app.AppSecret{AppID: otherApp.ID, Name: name, Value: "example"}
			err := s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
				return tx.Create(secret).Error
			}, &sql.TxOptions{Isolation: isolation})
			if isolation == sql.LevelReadCommitted {
				require.NoError(s.T(), err)
			} else {
				var pgErr *pgconn.PgError
				require.ErrorAs(s.T(), err, &pgErr)
				require.Equal(s.T(), "25000", pgErr.Code)
				require.Contains(s.T(), pgErr.Message, "require READ COMMITTED")
			}
			require.Equal(s.T(), before, s.head(), "even an unmapped source must enforce isolation")
			var count int64
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppSecret{}).Where(app.AppSecret{ID: secret.ID}).Count(&count).Error)
			if isolation == sql.LevelReadCommitted {
				require.EqualValues(s.T(), 1, count)
			} else {
				require.Zero(s.T(), count)
			}

			err = s.deps.DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
				return tx.Model(&app.App{ID: s.install.AppID}).Update("name", name).Error
			}, &sql.TxOptions{Isolation: isolation})
			if isolation == sql.LevelReadCommitted {
				require.NoError(s.T(), err)
				require.Equal(s.T(), before.Revision+1, s.head().Revision)
				require.Equal(s.T(), name, s.fresh().App.Name)
			} else {
				var pgErr *pgconn.PgError
				require.ErrorAs(s.T(), err, &pgErr)
				require.Equal(s.T(), "25000", pgErr.Code)
				require.Equal(s.T(), before, s.head())
				require.NotEqual(s.T(), name, s.fresh().App.Name)
			}
		})
	}
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

func (s *InstallStateIntegrationSuite) TestMixedMissingAndExistingHeadsDoNotDeadlock() {
	var application app.App
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.App{ID: s.install.AppID}).First(&application).Error)
	sibling := s.deps.Seeder.CreateInstall(s.ctx, s.T(), &application)
	low, high := s.install.ID, sibling.ID
	if low > high {
		low, high = high, low
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: low}).Delete(&app.InstallStateHead{}).Error)

	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	tx := s.deps.DB.WithContext(ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), s.deps.Helpers.WithDB(tx).MarkInstallStatePartialsStale(ctx, low, pkgstate.PartialInputs))
	pid := s.backendPID(tx)
	done := make(chan struct{})
	var fanoutErr error
	go func() {
		defer close(done)
		fanoutErr = s.deps.DB.WithContext(ctx).Exec(
			`SELECT invalidate_install_state(ARRAY[?, ?]::text[], '["app"]'::jsonb)`, high, low,
		).Error
	}()
	require.True(s.T(), s.blockedBy(pid, done), "fan-out must wait for the lower, newly created head")
	require.NoError(s.T(), tx.Exec(
		`SELECT invalidate_install_state(ARRAY[?, ?]::text[], '["cloud"]'::jsonb)`, high, low,
	).Error, "fan-out must not lock the existing higher head before the missing lower head")
	require.NoError(s.T(), tx.Commit().Error)
	s.await(done)
	require.NoError(s.T(), fanoutErr)

	for _, installID := range []string{low, high} {
		var head app.InstallStateHead
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: installID}).First(&head).Error)
		require.Contains(s.T(), head.StalePartials, pkgstate.PartialApp)
		require.Contains(s.T(), head.StalePartials, pkgstate.PartialCloud)
	}
}

func (s *InstallStateIntegrationSuite) TestInvalidationDDLTimesOutAndRetriesPerTable() {
	var blockedStage, freeStage func(context.Context, *gorm.DB) error
	for _, stage := range new(psqlmigrations.Migrations).All() {
		switch stage.Name {
		case "145-install-state-invalidation-runner_jobs":
			blockedStage = stage.Fn
		case "145-install-state-invalidation-install_inputs":
			freeStage = stage.Fn
		}
	}
	require.NotNil(s.T(), blockedStage)
	require.NotNil(s.T(), freeStage)
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()
	tx := s.deps.DB.WithContext(ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), tx.Exec("LOCK TABLE runner_jobs IN ROW EXCLUSIVE MODE").Error)
	require.ErrorContains(s.T(), blockedStage(ctx, s.deps.DB), "lock timeout")
	require.NoError(s.T(), freeStage(ctx, s.deps.DB), "a blocked runner table must not prevent another table's DDL from committing")
	require.NoError(s.T(), tx.Rollback().Error)
	require.NoError(s.T(), blockedStage(ctx, s.deps.DB))
	s.updateInput()
	require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"])
}

func (s *InstallStateIntegrationSuite) TestArchiveRetainsCurrentSnapshot() {
	oldest := s.currentRow()
	for n := range 55 {
		s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
			map[string]*string{"region": generics.ToPtr(fmt.Sprintf("region-%d", n))})
		require.Equal(s.T(), fmt.Sprintf("region-%d", n), s.fresh().Inputs.Inputs["region"])
	}
	var unarchived int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where(app.InstallState{InstallID: s.install.ID}).Where("NOT archived").Count(&unarchived).Error)
	require.Equal(s.T(), int64(50), unarchived, "read-driven publication must own retention without a state workflow")
	var historical app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{ID: oldest.ID}).First(&historical).Error)
	require.True(s.T(), historical.Archived)
	require.Nil(s.T(), historical.State)
	stored, fromBlob := historical.GetState(s.ctx, true)
	require.True(s.T(), fromBlob)
	require.Equal(s.T(), "us-east-1", stored.Inputs.Inputs["region"])

	current := s.currentRow()
	newer := make([]string, 0, 55)
	for range 55 {
		row := &app.InstallState{
			InstallID: s.install.ID,
			State:     &state.State{ID: s.install.ID, Name: "historical-snapshot"},
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
	require.Equal(s.T(), "region-54", snapshot.Inputs.Inputs["region"])
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
