package helpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/soft_delete"
	"hegel.dev/go/hegel"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type buildSourceRow struct {
	ID              string `gorm:"primaryKey"`
	DeletedAt       soft_delete.DeletedAt
	UpdatedAt       time.Time
	SourceIdentity  *sourceidentity.Identity `gorm:"type:jsonb;serializer:json"`
	Checksum        string
	SourceChecksum  string
	SourceRef       string
	SourceImage     string
	ResolvedTag     string
	SourceDigest    string
	SourceMediaType string
	ResolvedAt      *time.Time
	NoOp            bool
}

func buildSourcePostgres(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("initdb cannot create a disposable cluster as root")
	}
	initdb, err := exec.LookPath("initdb")
	if err != nil {
		t.Skip("local initdb is required for disposable PostgreSQL receipt tests")
	}
	pgctl, err := exec.LookPath("pg_ctl")
	if err != nil {
		t.Skip("local pg_ctl is required for disposable PostgreSQL receipt tests")
	}
	initdb, err = filepath.EvalSymlinks(initdb)
	require.NoError(t, err)
	pgctl, err = filepath.EvalSymlinks(pgctl)
	require.NoError(t, err)
	root := t.TempDir()
	socket, err := os.MkdirTemp("/tmp", "nd14pg-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(socket)) })
	data := filepath.Join(root, "data")
	out, err := exec.Command(initdb, "-D", data, "-A", "trust", "-U", "receipt_test", "--no-locale", "--encoding=UTF8", "--no-sync").CombinedOutput()
	require.NoError(t, err, "%s", out)
	out, err = exec.Command(pgctl, "-D", data, "-l", filepath.Join(root, "postgres.log"), "-o", fmt.Sprintf("-k %s -h '' -F", socket), "-w", "start").CombinedOutput()
	require.NoError(t, err, "%s", out)
	t.Cleanup(func() {
		out, err := exec.Command(pgctl, "-D", data, "-m", "immediate", "-w", "stop").CombinedOutput()
		require.NoError(t, err, "%s", out)
	})
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=receipt_test dbname=postgres sslmode=disable", socket)), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.Exec("CREATE EXTENSION hstore").Error)
	require.NoError(t, db.AutoMigrate(&app.RunnerJobExecutionResult{}))
	for _, owner := range []string{"component_builds", "app_sandbox_builds"} {
		require.NoError(t, db.Table(owner).AutoMigrate(&buildSourceRow{}))
	}
	require.NoError(t, db.Exec("CREATE TABLE runner_jobs (id text PRIMARY KEY, owner_type text, owner_id text, deleted_at bigint NOT NULL DEFAULT 0)").Error)
	require.NoError(t, db.Exec("CREATE TABLE runner_job_executions (id text PRIMARY KEY, runner_job_id text, deleted_at bigint NOT NULL DEFAULT 0)").Error)
	return db
}

func TestBuildSourcePostgresReceiptLifecycle(t *testing.T) {
	db := buildSourcePostgres(t)
	for _, owner := range []string{"component_builds", "app_sandbox_builds"} {
		t.Run(owner, func(t *testing.T) {
			hegel.Test(t, func(t *hegel.T) {
				defer func() {
					require.NoError(t, db.Exec("TRUNCATE component_builds, app_sandbox_builds, runner_jobs, runner_job_executions, runner_job_execution_results").Error)
				}()
				prepared := drawBuildSource(t)
				prepared.Captured = hegel.Draw(t, hegel.Booleans())
				if hegel.Draw(t, hegel.Booleans()) {
					prepared.Coverage, prepared.Reasons = sourceidentity.Complete, nil
				}
				if !prepared.Captured {
					prepared.Digest = ""
					prepared.MarkUnknown("source_requires_runner_resolution")
				}
				prepared.Seal()
				before := cloneBuildSource(prepared)
				claimed := cloneBuildSource(prepared)
				claimed.Captured = true
				if !prepared.Captured {
					if claimed.Kind == "image" {
						claimed.Digest = "sha256:" + sourceidentity.Hash("resolved")
						claimed.Reasons = slices.DeleteFunc(claimed.Reasons, func(s string) bool { return s == "source_requires_runner_resolution" })
						if len(claimed.Reasons) == 0 {
							claimed.Coverage, claimed.Reasons = sourceidentity.Complete, nil
						}
					} else {
						claimed.MarkUnknown("dependency_closure_uninspected")
					}
				}
				claimed.Seal()
				job := &app.RunnerJob{ID: strings.Repeat("j", 26), OwnerID: strings.Repeat("b", 26), OwnerType: owner}
				execution := strings.Repeat("e", 26)
				row := &buildSourceRow{ID: job.OwnerID, SourceIdentity: prepared, Checksum: prepared.ConfigChecksum}
				require.NoError(t, db.Table(owner).Create(row).Error)
				require.NoError(t, db.Table("runner_jobs").Create(map[string]any{"id": job.ID, "owner_id": job.OwnerID, "owner_type": owner}).Error)
				require.NoError(t, db.Table("runner_job_executions").Create(map[string]any{"id": execution, "runner_job_id": job.ID}).Error)
				read := func() buildSourceRow {
					var stored buildSourceRow
					require.NoError(t, db.Table(owner).Where(buildSourceRow{ID: job.OwnerID}).First(&stored).Error)
					return stored
				}
				ctx := context.WithValue(context.Background(), keys.AccountIDCtxKey, strings.Repeat("a", 26))
				require.Error(t, ClaimBuildSource(ctx, db, job, "missing", claimed))
				require.Equal(t, before, read().SourceIdentity)
				require.NoError(t, ClaimBuildSource(ctx, db, job, execution, claimed))
				require.Equal(t, claimed, read().SourceIdentity)
				require.Empty(t, read().SourceChecksum)
				plan := &plantypes.CompositePlan{BuildPlan: &plantypes.BuildPlan{SourceIdentity: prepared}}
				resumed, err := ResumeBuildSource(ctx, db, job.ID, plan)
				require.NoError(t, err)
				require.Equal(t, claimed, resumed.BuildPlan.SourceIdentity)
				require.Equal(t, before, prepared)
				changed := cloneBuildSource(claimed)
				changed.Files[0].Executable = !changed.Files[0].Executable
				changed.Seal()
				require.Error(t, ClaimBuildSource(ctx, db, job, execution, changed))
				badPlan := &plantypes.CompositePlan{BuildPlan: &plantypes.BuildPlan{SourceIdentity: changed}}
				_, err = ResumeBuildSource(ctx, db, job.ID, badPlan)
				require.Error(t, err)
				failed := &app.RunnerJobExecutionResult{RunnerJobExecutionID: strings.Repeat("f", 26), Success: false}
				_, created, err := PersistBuildResult(ctx, db, job, failed, &app.ComponentBuild{SourceIdentity: changed})
				require.NoError(t, err)
				require.True(t, created)
				require.Equal(t, claimed, read().SourceIdentity)
				result := &app.RunnerJobExecutionResult{RunnerJobExecutionID: execution, Success: true}
				_, _, err = PersistBuildResult(ctx, db, job, result, &app.ComponentBuild{SourceIdentity: changed})
				require.Error(t, err)
				var count int64
				require.NoError(t, db.Model(&app.RunnerJobExecutionResult{}).Where(app.RunnerJobExecutionResult{RunnerJobExecutionID: execution}).Count(&count).Error)
				require.Zero(t, count)
				source := &app.ComponentBuild{SourceIdentity: claimed}
				if owner == "component_builds" && claimed.Kind == "image" {
					source.SourceDigest = "sha256:" + sourceidentity.Hash(claimed.Digest)
					_, _, err = PersistBuildResult(ctx, db, job, result, source)
					require.Error(t, err)
					require.Equal(t, claimed, read().SourceIdentity)
					source.SourceDigest = claimed.Digest
				}
				persisted, created, err := PersistBuildResult(ctx, db, job, result, source)
				require.NoError(t, err)
				require.True(t, created)
				wantEligible := ""
				if claimed.Coverage == sourceidentity.Complete {
					wantEligible = claimed.Checksum
				}
				require.Equal(t, wantEligible, read().SourceChecksum)
				require.Equal(t, claimed, read().SourceIdentity)
				replay := &app.RunnerJobExecutionResult{RunnerJobExecutionID: execution, Success: false}
				replayed, created, err := PersistBuildResult(ctx, db, job, replay, &app.ComponentBuild{SourceIdentity: changed})
				require.NoError(t, err)
				require.False(t, created)
				require.Equal(t, persisted.ID, replayed.ID)
				require.True(t, replayed.Success)
				require.Equal(t, claimed, read().SourceIdentity)
				missing := &app.RunnerJobExecutionResult{RunnerJobExecutionID: strings.Repeat("m", 26), Success: true}
				_, _, err = PersistBuildResult(ctx, db, job, missing, &app.ComponentBuild{})
				require.NoError(t, err)
				stored := read()
				require.Empty(t, stored.SourceChecksum)
				require.Equal(t, claimed.Checksum, stored.SourceIdentity.Checksum)
				require.Contains(t, stored.SourceIdentity.Reasons, "runner_source_evidence_missing")
				for _, reason := range claimed.Reasons {
					require.Contains(t, stored.SourceIdentity.Reasons, reason)
				}
				require.Equal(t, before, prepared)
			}, hegel.WithTestCases(30))
		})
	}
}

func TestBuildSourcePostgresConcurrentClaimsDoNotRepin(t *testing.T) {
	db := buildSourcePostgres(t)
	hegel.Test(t, func(t *hegel.T) {
		defer func() {
			require.NoError(t, db.Exec("TRUNCATE component_builds, runner_job_executions").Error)
		}()
		prepared := drawBuildSource(t)
		prepared.Kind, prepared.Digest = "image", ""
		prepared.Seal()
		before := cloneBuildSource(prepared)
		first, second := cloneBuildSource(prepared), cloneBuildSource(prepared)
		first.Captured, second.Captured = true, true
		first.Digest = "sha256:" + sourceidentity.Hash(hegel.Draw(t, hegel.Text().MaxSize(16)))
		second.Digest = "sha256:" + sourceidentity.Hash(first.Digest)
		first.Seal()
		second.Seal()
		job := &app.RunnerJob{ID: strings.Repeat("j", 26), OwnerID: strings.Repeat("b", 26), OwnerType: "component_builds"}
		execution := strings.Repeat("e", 26)
		require.NoError(t, db.Table("component_builds").Create(&buildSourceRow{ID: job.OwnerID, SourceIdentity: prepared}).Error)
		require.NoError(t, db.Table("runner_job_executions").Create(map[string]any{"id": execution, "runner_job_id": job.ID}).Error)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		start := make(chan struct{})
		outcomes := make(chan struct {
			identity *sourceidentity.Identity
			err      error
		}, 2)
		for _, identity := range []*sourceidentity.Identity{first, second} {
			go func() {
				<-start
				err := ClaimBuildSource(ctx, db, job, execution, identity)
				outcomes <- struct {
					identity *sourceidentity.Identity
					err      error
				}{identity, err}
			}()
		}
		close(start)
		a, b := <-outcomes, <-outcomes
		require.NotEqual(t, a.err == nil, b.err == nil)
		winner, loser := a.identity, b.identity
		if a.err != nil {
			winner, loser = loser, winner
		}
		var stored buildSourceRow
		require.NoError(t, db.Table("component_builds").Where(buildSourceRow{ID: job.OwnerID}).First(&stored).Error)
		require.Equal(t, winner, stored.SourceIdentity)
		require.NoError(t, ClaimBuildSource(ctx, db, job, execution, winner))
		require.Error(t, ClaimBuildSource(ctx, db, job, execution, loser))
		require.Empty(t, stored.SourceChecksum)
		require.Equal(t, before, prepared)
	}, hegel.WithTestCases(20))
}
