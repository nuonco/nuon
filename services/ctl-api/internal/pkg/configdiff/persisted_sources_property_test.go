package configdiff

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/afterquery"
)

func isolatedConfigDiffDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("CONFIGDIFF_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("CONFIGDIFF_TEST_POSTGRES_DSN is not set")
	}
	cfg, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	require.Contains(t, []string{"localhost", "127.0.0.1", "::1"}, cfg.Host)
	for _, fallback := range cfg.Fallbacks {
		require.Contains(t, []string{"localhost", "127.0.0.1", "::1"}, fallback.Host)
	}
	cfg.Database = "postgres"
	cfg.ConnectTimeout = 5 * time.Second
	admin := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	name := "configdiff_pbt_" + strings.ToLower(rand.Text())
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP DATABASE " + name)
		require.NoError(t, err)
	})
	cfg.Database = name
	conn := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Use(afterquery.NewAfterQueryPlugin()))
	require.NoError(t, db.Exec("CREATE EXTENSION hstore").Error)
	require.NoError(t, db.AutoMigrate(
		&app.AppConfig{}, &app.Component{}, &app.ComponentConfigConnection{},
		&app.ComponentBuild{}, &app.AppSandboxBuild{}, &app.AppSandboxConfig{},
		&app.AppStackConfig{}, &app.AppRunnerConfig{}, &app.AppPermissionsConfig{},
		&app.AppBreakGlassConfig{}, &app.AppSecretsConfig{},
		&app.PublicGitVCSConfig{}, &app.ConnectedGithubVCSConfig{},
	))
	return db
}

func persistedFactID(group, index int) string {
	return fmt.Sprintf("%013d%013d", group, index)
}

type persistedSourceFact struct {
	known   bool
	content int
}

func (f persistedSourceFact) buildFact() buildSourceFact {
	evidence := "unknown"
	if f.known {
		evidence = "complete"
	}
	return buildSourceFact{evidence: evidence, content: f.content}
}

type persistedSourceModel struct {
	sources       [5][5]persistedSourceFact
	configValues  [3][4]int
	baselines     [4]int
	buildOverride [4]int
	receipts      [4]int
	sandbox       int
	sandboxBuild  int
	buildMap      bool
	intermediate  bool
	dependencyRef bool
	sandboxPinned bool
}

func (m persistedSourceModel) equal(a, b, resource int) bool {
	return a >= 0 && b >= 0 && (a == b ||
		(m.sources[a][resource].known && m.sources[b][resource].known &&
			m.sources[a][resource].content == m.sources[b][resource].content))
}

func (m persistedSourceModel) baselineBuild(version int, later bool) int {
	if version == 0 && later {
		return 4
	}
	return version
}

func (m persistedSourceModel) oldBuild(i int, later bool) int {
	if m.buildMap && m.buildOverride[i] != -2 {
		return m.buildOverride[i]
	}
	return m.baselineBuild(m.baselines[i], later)
}

func (m persistedSourceModel) edges() [][2]int {
	return [][2]int{{1, 0}, {2, 4}, {3, 1}}
}

func (m persistedSourceModel) impacts(i int, later bool) []pkgdiff.ImpactReason {
	var direct, reached [2][5]bool
	version := m.baselines[i]
	for resource := range 5 {
		old := m.baselineBuild(version, later && (resource != 4 || !m.sandboxPinned))
		if m.buildMap || m.receipts[i] != -2 {
			old = m.receipts[i]
		}
		if resource == i {
			old = m.oldBuild(i, later)
		}
		direct[0][resource] = resource < 4 && m.configValues[version][resource] != m.configValues[2][resource]
		direct[1][resource] = direct[0][resource] || !m.equal(old, 2, resource)
	}
	reached = direct
	var reasons []pkgdiff.ImpactReason
	for stage := range 2 {
		if stage == 0 && !m.intermediate {
			continue
		}
		for range 5 {
			for _, edge := range m.edges() {
				if reached[stage][edge[1]] {
					reached[stage][edge[0]] = true
				}
			}
		}
		if direct[stage][i] {
			continue
		}
		var nextReasons []pkgdiff.ImpactReason
		for _, edge := range m.edges() {
			if edge[0] == i && reached[stage][edge[1]] {
				reason := pkgdiff.EdgeReasonComponentDependency
				if edge[1] == 4 {
					reason = pkgdiff.EdgeReasonSandboxOutput
				} else if i == 1 && m.dependencyRef {
					reason = pkgdiff.EdgeReasonComponentReference
				}
				nextReasons = append(nextReasons, pkgdiff.ImpactReason{From: persistedResource(edge[1]), Edge: reason})
			}
		}
		if len(nextReasons) > 0 {
			reasons = nextReasons
		}
	}
	return reasons
}

func persistedResource(i int) pkgdiff.NodeID {
	if i == 4 {
		return "sandbox"
	}
	return pkgdiff.NodeID(fmt.Sprintf("component.resource-%d", i))
}

type configDiffBlobs struct {
	blobstore.Service
	values map[string]string
}

func (b configDiffBlobs) DownloadStream(_ context.Context, key string) (io.ReadCloser, error) {
	value, ok := b.values[key]
	if !ok {
		return nil, fmt.Errorf("missing test blob %s", key)
	}
	return io.NopCloser(strings.NewReader(value)), nil
}

func seedPersistedSourceModel(t require.TestingT, db *gorm.DB, m persistedSourceModel) (context.Context, CompositeBaselines) {
	write := db.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Session(&gorm.Session{})
	creator, org, appID := persistedFactID(1, 0), persistedFactID(2, 0), persistedFactID(3, 0)
	blobs := configDiffBlobs{values: map[string]string{}}
	for i := range 4 {
		require.NoError(t, write.Create(&app.Component{
			ID: persistedFactID(10, i), CreatedByID: creator, OrgID: org,
			Name: fmt.Sprintf("resource-%d", i), Type: app.ComponentTypeJob,
		}).Error)
	}
	for version := range 3 {
		cfgID := persistedFactID(20, version)
		cfg := &app.AppConfig{
			ID: cfgID, CreatedByID: creator, OrgID: org, AppID: appID,
			BuildsPinned: m.sandboxPinned, BuildsLocked: m.sandboxPinned,
			SandboxBuildID: generics.NewNullString(persistedFactID(60, version)),
		}
		intermediate := &config.AppConfig{Sandbox: &config.AppSandboxConfig{TerraformVersion: "1.7.0"}}
		for i := range 4 {
			conn := &app.ComponentConfigConnection{
				ID: persistedFactID(30+version, i), CreatedByID: creator, OrgID: org,
				AppConfigID: cfgID, ComponentID: persistedFactID(10, i),
				Checksum:      fmt.Sprintf("config-%d-%d", i, m.configValues[version][i]),
				LatestBuildID: generics.NewNullString(persistedFactID(40+version, i)),
			}
			component := &config.Component{
				Name: fmt.Sprintf("resource-%d", i), Type: config.JobComponentType,
				VarName: conn.Checksum, Job: &config.JobComponentConfig{ImageURL: "example.com/acme/job", Tag: "v1"},
			}
			for _, edge := range m.edges() {
				if edge[0] != i {
					continue
				}
				if edge[1] == 4 {
					conn.References = append(conn.References, "sandbox||output|")
					component.References = append(component.References, refs.NewFromString("sandbox||output|"))
				} else if i == 1 && m.dependencyRef {
					ref := fmt.Sprintf("component|resource-%d|output|", edge[1])
					conn.References = append(conn.References, ref)
					component.References = append(component.References, refs.NewFromString(ref))
				} else {
					conn.ComponentDependencyIDs = append(conn.ComponentDependencyIDs, persistedFactID(10, edge[1]))
					component.Dependencies = append(component.Dependencies, fmt.Sprintf("resource-%d", edge[1]))
				}
			}
			require.NoError(t, write.Create(conn).Error)
			intermediate.Components = append(intermediate.Components, component)
		}
		if m.intermediate {
			raw, err := json.Marshal(intermediate)
			require.NoError(t, err)
			blobs.values[cfgID] = string(raw)
			metadata, err := json.Marshal(blobstore.BlobMetadata{BlobID: cfgID, S3Key: cfgID})
			require.NoError(t, err)
			cfg.IntermediateConfig = &blobstore.Blob{}
			require.NoError(t, cfg.IntermediateConfig.Scan(metadata))
		}
		require.NoError(t, write.Create(cfg).Error)
		require.NoError(t, write.Create(&app.AppSandboxConfig{
			ID: persistedFactID(50, version), CreatedByID: creator, OrgID: org, AppID: appID,
			AppConfigID: cfgID, Type: "terraform", TerraformVersion: "1.7.0",
		}).Error)
	}
	for slot := range 4 {
		seedPersistedBuildSlot(t, write, m, slot)
	}
	if m.sandboxPinned {
		seedPersistedBuildSlot(t, write, m, 4)
	}
	baselines := CompositeBaselines{
		Fallback: persistedFactID(20, 0), Stack: persistedFactID(20, 0),
		Sandbox: persistedFactID(20, m.sandbox), Components: map[string]string{},
		DependencyBuildIDs: map[string]map[string]string{},
	}
	if m.buildMap {
		baselines.ComponentBuilds = map[string]*app.ComponentBuild{}
	}
	for i := range 4 {
		id := persistedFactID(10, i)
		if m.baselines[i] != 0 {
			baselines.Components[id] = ""
			if m.baselines[i] >= 0 {
				baselines.Components[id] = persistedFactID(20, m.baselines[i])
			} else if m.baselines[i] == -3 {
				baselines.Components[id] = persistedFactID(20, 99)
			}
		}
		if m.buildMap && m.buildOverride[i] != -2 {
			baselines.ComponentBuilds[id] = nil
			if m.buildOverride[i] >= 0 {
				var build app.ComponentBuild
				require.NoError(t, db.Unscoped().Where(app.ComponentBuild{ID: persistedFactID(40+m.buildOverride[i], i)}).First(&build).Error)
				baselines.ComponentBuilds[id] = &build
			} else if m.buildOverride[i] == -3 {
				baselines.ComponentBuilds[id] = &app.ComponentBuild{}
			}
		}
		if m.receipts[i] != -2 {
			receipt := map[string]string{}
			if m.receipts[i] >= 0 || m.receipts[i] <= -3 {
				slot := m.receipts[i]
				if slot < 0 {
					slot = 99
				}
				for resource := range 4 {
					receipt[persistedFactID(10, resource)] = persistedFactID(40+slot, resource)
					if m.receipts[i] == -4 {
						receipt[persistedFactID(10, resource)] = ""
					}
				}
				receipt["sandbox"] = persistedFactID(60, slot)
				if m.receipts[i] == -4 {
					receipt["sandbox"] = ""
				}
			}
			baselines.DependencyBuildIDs[id] = receipt
		}
	}
	if m.sandboxBuild != -2 {
		id := ""
		if m.sandboxBuild >= 0 {
			id = persistedFactID(60, m.sandboxBuild)
		}
		baselines.SandboxBuildID = &id
	}
	return blobstore.WithBlobService(context.Background(), blobs), baselines
}

func seedPersistedBuildSlot(t require.TestingT, db *gorm.DB, m persistedSourceModel, slot int) {
	version := min(slot, 2)
	if slot >= 3 {
		version = 0
	}
	for i := range 4 {
		build := m.sources[slot][i].buildFact().component(persistedFactID(40+slot, i))
		build.CreatedByID, build.OrgID = persistedFactID(1, 0), persistedFactID(2, 0)
		build.ComponentConfigConnectionID = persistedFactID(30+version, i)
		build.Status = app.ComponentBuildStatusActive
		build.CreatedAt = time.Unix(int64(100+slot), 0)
		if slot == 3 {
			build.DeletedAt = 1
		}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(build).Error)
	}
	build := m.sources[slot][4].buildFact().sandbox(persistedFactID(60, slot))
	build.CreatedByID, build.OrgID = persistedFactID(1, 0), persistedFactID(2, 0)
	build.AppID, build.AppConfigID = persistedFactID(3, 0), persistedFactID(20, version)
	build.AppSandboxConfigID = persistedFactID(50, version)
	build.Status = app.AppSandboxBuildStatusActive
	build.CreatedAt = time.Unix(int64(100+slot), 0)
	if slot == 3 {
		build.DeletedAt = 1
	}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(build).Error)
}

func requirePersistedSourceModel(t require.TestingT, m persistedSourceModel, diff *app.InstallConfigDiff, later bool) {
	want := map[string][]string{"added": {}, "changed": {}, "unchanged": {}, "removed": {}}
	entries := map[string]app.ComponentDiffEntry{}
	for category, got := range map[string][]app.ComponentDiffEntry{
		"added": diff.Added, "changed": diff.Changed, "unchanged": diff.Unchanged, "removed": diff.Removed,
	} {
		for _, entry := range got {
			_, duplicate := entries[entry.ComponentID]
			require.False(t, duplicate)
			entries[entry.ComponentID] = entry
			want[category] = append(want[category], entry.ComponentID)
		}
	}
	require.Len(t, entries, 4)
	expected := map[string][]string{"added": {}, "changed": {}, "unchanged": {}, "removed": {}}
	for i := range 4 {
		id := persistedFactID(10, i)
		entry := entries[id]
		require.Equal(t, fmt.Sprintf("resource-%d", i), entry.ComponentName)
		category, oldID := "added", ""
		if m.baselines[i] >= 0 {
			old := m.oldBuild(i, later)
			if old >= 0 {
				oldID = persistedFactID(40+old, i)
			}
			reasons := m.impacts(i, later)
			require.ElementsMatch(t, reasons, entry.ImpactReasons, "resource=%d model=%+v", i, m)
			category = "unchanged"
			if m.configValues[m.baselines[i]][i] != m.configValues[2][i] || !m.equal(old, 2, i) || len(reasons) > 0 {
				category = "changed"
			}
		}
		expected[category] = append(expected[category], id)
		require.Equal(t, oldID, entry.OldBuildID, "resource=%d", i)
		require.Equal(t, persistedFactID(42, i), entry.NewBuildID)
		require.Equal(t, oldID != entry.NewBuildID, entry.BuildChanged)
	}
	require.Equal(t, expected, want, "model=%+v later=%t", m, later)
	old := m.sandboxBuild
	if old == -2 {
		old = m.baselineBuild(m.sandbox, later && !m.sandboxPinned)
	}
	oldID := ""
	if old >= 0 {
		oldID = persistedFactID(60, old)
	}
	require.Equal(t, oldID, diff.SandboxBuildOldID)
	require.Equal(t, persistedFactID(60, 2), diff.SandboxBuildNewID)
	require.Equal(t, old != 2, diff.SandboxBuildChanged)
	require.Equal(t, !m.equal(old, 2, 4), diff.SandboxChanged)
}

func TestCompositePersistedSourceSelectionProperties(t *testing.T) {
	t.Parallel()
	db := isolatedConfigDiffDB(t)
	hegel.Test(t, func(t *hegel.T) {
		var m persistedSourceModel
		m.buildMap = hegel.Draw(t, hegel.Booleans())
		m.intermediate = hegel.Draw(t, hegel.Booleans())
		m.dependencyRef = hegel.Draw(t, hegel.Booleans())
		m.sandboxPinned = hegel.Draw(t, hegel.Booleans())
		m.sandbox = hegel.Draw(t, hegel.Integers(0, 1))
		m.sandboxBuild = hegel.Draw(t, hegel.SampledFrom([]int{-2, -1, 0, 2, 3}))
		for slot := range 5 {
			for resource := range 5 {
				m.sources[slot][resource] = persistedSourceFact{
					known: hegel.Draw(t, hegel.Booleans()), content: hegel.Draw(t, hegel.Integers(0, 1)),
				}
			}
		}
		for version := range 3 {
			for i := range 4 {
				m.configValues[version][i] = hegel.Draw(t, hegel.Integers(0, 1))
			}
		}
		for i := range 4 {
			m.baselines[i] = hegel.Draw(t, hegel.SampledFrom([]int{-3, -1, 0, 1, 2}))
			m.buildOverride[i] = hegel.Draw(t, hegel.SampledFrom([]int{-3, -2, -1, 0, 2, 3}))
			m.receipts[i] = hegel.Draw(t, hegel.SampledFrom([]int{-4, -3, -2, -1, 0, 2, 3}))
		}
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		ctx, baselines := seedPersistedSourceModel(t, tx, m)
		before, err := json.Marshal(baselines)
		require.NoError(t, err)
		diff, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
		require.NoError(t, err)
		requirePersistedSourceModel(t, m, diff, false)
		again, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
		require.NoError(t, err)
		require.Equal(t, diff, again)
		after, err := json.Marshal(baselines)
		require.NoError(t, err)
		require.Equal(t, before, after)
	}, hegel.WithTestCases(40))
}

func TestCompositeConsumedSourcesIgnoreLaterLatestProperties(t *testing.T) {
	t.Parallel()
	db := isolatedConfigDiffDB(t)
	hegel.Test(t, func(t *hegel.T) {
		m := persistedSourceModel{buildMap: true, sandboxBuild: 3}
		m.intermediate = hegel.Draw(t, hegel.Booleans())
		m.dependencyRef = hegel.Draw(t, hegel.Booleans())
		for resource := range 5 {
			content := hegel.Draw(t, hegel.Integers(0, 1))
			for slot := range 5 {
				m.sources[slot][resource] = persistedSourceFact{known: true, content: content}
			}
			m.sources[0][resource].content = 1 - content
			m.sources[4][resource] = persistedSourceFact{known: hegel.Draw(t, hegel.Booleans()), content: 1 - content}
		}
		for i := range 4 {
			m.baselines[i] = hegel.Draw(t, hegel.Integers(0, 1))
			m.buildOverride[i], m.receipts[i] = 3, 3
		}
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		ctx, baselines := seedPersistedSourceModel(t, tx, m)
		before, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
		require.NoError(t, err)
		requirePersistedSourceModel(t, m, before, false)
		require.Len(t, before.Unchanged, 4)
		seedPersistedBuildSlot(t, tx, m, 4)
		for i := range 4 {
			require.NoError(t, tx.Model(&app.ComponentConfigConnection{}).
				Where(app.ComponentConfigConnection{ID: persistedFactID(30, i)}).
				Update("latest_build_id", persistedFactID(44, i)).Error)
		}
		after, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
		require.NoError(t, err)
		requirePersistedSourceModel(t, m, after, true)
		require.Equal(t, before, after)
		m.buildMap, m.sandboxBuild = false, -2
		baselines.ComponentBuilds, baselines.SandboxBuildID = nil, nil
		for i := range 4 {
			m.receipts[i] = -2
		}
		baselines.DependencyBuildIDs = nil
		control, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
		require.NoError(t, err)
		requirePersistedSourceModel(t, m, control, true)
		require.True(t, control.SandboxChanged)
	}, hegel.WithTestCases(24))
}

func TestCompositeReceiptIsolationAndEmptyOverridesProperties(t *testing.T) {
	t.Parallel()
	db := isolatedConfigDiffDB(t)
	hegel.Test(t, func(t *hegel.T) {
		base := persistedSourceModel{buildMap: true, sandboxBuild: 2}
		base.intermediate = hegel.Draw(t, hegel.Booleans())
		base.dependencyRef = hegel.Draw(t, hegel.Booleans())
		base.sandboxPinned = hegel.Draw(t, hegel.Booleans())
		for resource := range 5 {
			content := hegel.Draw(t, hegel.Integers(0, 1))
			for slot := range 5 {
				base.sources[slot][resource] = persistedSourceFact{known: true, content: content}
			}
		}
		base.sources[3][0].content = 1 - base.sources[2][0].content
		base.sources[3][4].content = 1 - base.sources[2][4].content
		for i := range 4 {
			base.buildOverride[i], base.receipts[i] = 2, 2
		}
		for _, mode := range []string{
			"isolated_receipts", "absent_maps", "empty_build_map", "empty_receipts",
			"nil_build", "empty_build", "empty_baseline", "missing_baseline", "empty_sandbox",
			"missing_receipt_rows", "empty_receipt_values", "authored_impact_with_unknown_source", "missing_sandbox_build",
		} {
			t.Log(mode)
			m := base
			switch mode {
			case "isolated_receipts":
				m.receipts[1], m.receipts[2] = 3, 3
			case "absent_maps":
				m.buildMap, m.sandboxBuild = false, -2
				for i := range 4 {
					m.receipts[i] = -2
				}
			case "empty_build_map":
				for i := range 4 {
					m.buildOverride[i], m.receipts[i] = -2, -2
				}
			case "empty_receipts":
				for i := range 4 {
					m.receipts[i] = -1
				}
			case "missing_receipt_rows":
				for i := range 4 {
					m.receipts[i] = -3
				}
			case "empty_receipt_values":
				for i := range 4 {
					m.receipts[i] = -4
				}
			case "nil_build":
				m.buildOverride[0] = -1
			case "empty_build":
				m.buildOverride[0] = -3
			case "empty_baseline":
				m.baselines[1] = -1
			case "missing_baseline":
				m.baselines[1] = -3
			case "empty_sandbox":
				m.sandboxBuild = -1
			case "authored_impact_with_unknown_source":
				m.intermediate, m.buildMap = true, false
				m.configValues[2][1] = 1
				for slot := range 5 {
					for resource := range 5 {
						m.sources[slot][resource].known = false
					}
				}
				for i := range 4 {
					m.receipts[i] = -2
				}
			}
			func() {
				tx := db.Begin()
				require.NoError(t, tx.Error)
				defer tx.Rollback()
				ctx, baselines := seedPersistedSourceModel(t, tx, m)
				if mode == "missing_sandbox_build" {
					id := persistedFactID(60, 99)
					baselines.SandboxBuildID = &id
				}
				diff, err := ComputeCompositeInstallConfigDiff(ctx, tx, baselines, persistedFactID(20, 2))
				if mode == "missing_sandbox_build" {
					require.ErrorContains(t, err, "unable to get applied sandbox source")
					require.Nil(t, diff)
					return
				}
				require.NoError(t, err)
				requirePersistedSourceModel(t, m, diff, false)
			}()
		}
	}, hegel.WithTestCases(8))
}
