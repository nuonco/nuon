package activities

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type registryPruneTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestGetRegistryPruneTagsSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(registryPruneTestSuite))
}

func (s *registryPruneTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	cfg, err := tests.LoadDBConfig()
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	s.seeder = testseed.New(testseed.Params{DB: s.db})
	s.SetDB(s.db)
}

func (s *registryPruneTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

type pruneDeploySpec struct {
	component int
	status    app.InstallDeployStatus
	age       time.Duration
	digest    string
	resolved  string
	plan      string
	noPlan    bool
	wantTag   string
}

type pruneCase struct {
	appMode       *string
	appRetain     *int
	appGraceDays  *int
	installMode   *string
	componentName []string
	compRetain    map[int]int
	deploys       []pruneDeploySpec
	req           GetRegistryPruneTagsRequest

	wantErr      string
	wantMode     string
	wantGrace    *time.Duration
	wantNoTags   bool
	wantOwned    []int
	wantKeepTags []int
	wantKeepIDs  []int
}

func ptr[T any](v T) *T { return &v }

func (s *registryPruneTestSuite) TestGetRegistryPruneTags() {
	const (
		active          = app.InstallDeployStatusActive
		errored         = app.InstallDeployStatusError
		pendingApproval = app.InstallDeployStatusPendingApproval
	)

	cases := map[string]pruneCase{
		"mode off returns early without tags": {
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: 2 * time.Hour},
				{status: active, age: time.Hour},
			},
			wantMode:   "off",
			wantNoTags: true,
		},
		"mode off with IncludeWhenOff computes tags": {
			componentName: []string{"alpha"},
			req:           GetRegistryPruneTagsRequest{IncludeWhenOff: true},
			deploys: []pruneDeploySpec{
				{status: active, age: 4 * time.Hour},
				{status: active, age: 3 * time.Hour},
				{status: active, age: 2 * time.Hour},
				{status: active, age: time.Hour},
			},
			wantMode:     "off",
			wantOwned:    []int{0, 1, 2, 3},
			wantKeepTags: []int{1, 2, 3},
			wantKeepIDs:  []int{3},
		},
		"install override enforce beats app dry_run": {
			appMode:       ptr("dry_run"),
			installMode:   ptr("enforce"),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: time.Hour},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0},
			wantKeepTags: []int{0},
			wantKeepIDs:  []int{0},
		},
		"install override off beats app enforce": {
			appMode:       ptr("enforce"),
			installMode:   ptr("off"),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: time.Hour},
			},
			wantMode:   "off",
			wantNoTags: true,
		},
		"app retain count keeps newest two plus live": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(2),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: 4 * time.Hour},
				{status: active, age: 3 * time.Hour},
				{status: active, age: 2 * time.Hour},
				{status: active, age: time.Hour},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2, 3},
			wantKeepTags: []int{2, 3},
			wantKeepIDs:  []int{3},
		},
		"component retain count overrides app": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(3),
			componentName: []string{"alpha", "beta"},
			compRetain:    map[int]int{0: 1},
			deploys: []pruneDeploySpec{
				{component: 0, status: active, age: 4 * time.Hour},
				{component: 0, status: active, age: 3 * time.Hour},
				{component: 0, status: active, age: 2 * time.Hour},
				{component: 1, status: active, age: 3 * time.Hour},
				{component: 1, status: active, age: 2 * time.Hour},
				{component: 1, status: active, age: time.Hour},
				{component: 1, status: active, age: 30 * time.Minute},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2, 3, 4, 5, 6},
			wantKeepTags: []int{2, 4, 5, 6},
			wantKeepIDs:  []int{2, 6},
		},
		"tag pinned in live plan is kept past retain count": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			componentName: []string{"img-whoami"},
			req:           GetRegistryPruneTagsRequest{SharedRepo: true},
			deploys: []pruneDeploySpec{
				{status: active, age: 3 * time.Hour, digest: "sha256:aaa", resolved: "v1.0.0", wantTag: "img-whoami-v1.0.0"},
				{status: active, age: 2 * time.Hour, digest: "sha256:bbb", resolved: "v1.1.0", wantTag: "img-whoami-v1.1.0"},
				{status: active, age: time.Hour, digest: "sha256:ccc", resolved: "v1.2.0", wantTag: "img-whoami-v1.2.0",
					plan: `{"helm":{"values":{"image":{"tag":"img-whoami-v1.0.0"}}}}`},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{0, 2},
			wantKeepIDs:  []int{2},
		},
		"in-flight deploy tag is kept past retain count": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: pendingApproval, age: 10 * time.Hour},
				{status: active, age: 3 * time.Hour},
				{status: active, age: time.Hour},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{0, 2},
			wantKeepIDs:  []int{0, 2},
		},
		"shared repo uses component segment and resolved tag": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(10),
			componentName: []string{"Whoami Svc"},
			req:           GetRegistryPruneTagsRequest{SharedRepo: true},
			deploys: []pruneDeploySpec{
				{status: active, age: 3 * time.Hour, digest: "sha256:aaa", resolved: "v2", wantTag: "whoami-svc-v2"},
				{status: active, age: 2 * time.Hour},
				{status: active, age: time.Hour, digest: "sha256:ccc"},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{0, 1, 2},
			wantKeepIDs:  []int{2},
		},
		"non-shared repo uses resolved tag alone": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(10),
			componentName: []string{"Whoami Svc"},
			deploys: []pruneDeploySpec{
				{status: active, age: time.Hour, digest: "sha256:aaa", resolved: "v2", wantTag: "v2"},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0},
			wantKeepTags: []int{0},
			wantKeepIDs:  []int{0},
		},
		"live deploy without saved plan errors": {
			appMode:       ptr("enforce"),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: time.Hour, noPlan: true},
			},
			wantErr: "has no saved plan",
		},
		"grace period keeps recent deploys": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			componentName: []string{"alpha"},
			req:           GetRegistryPruneTagsRequest{GracePeriodOverride: ptr(time.Hour)},
			deploys: []pruneDeploySpec{
				{status: active, age: 5 * time.Hour},
				{status: errored, age: 20 * time.Minute},
				{status: active, age: 10 * time.Minute},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{1, 2},
			wantKeepIDs:  []int{1, 2},
		},
		"default grace period is 30 days": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: 40 * 24 * time.Hour},
				{status: errored, age: 20 * 24 * time.Hour},
				{status: active, age: time.Hour},
			},
			wantMode:     "enforce",
			wantGrace:    ptr(30 * 24 * time.Hour),
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{1, 2},
			wantKeepIDs:  []int{1, 2},
		},
		"app grace_period_days keeps recent deploys": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			appGraceDays:  ptr(1),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: 48 * time.Hour},
				{status: errored, age: 20 * time.Hour},
				{status: active, age: 10 * time.Minute},
			},
			wantMode:     "enforce",
			wantGrace:    ptr(24 * time.Hour),
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{1, 2},
			wantKeepIDs:  []int{1, 2},
		},
		"admin override beats app grace_period_days": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			appGraceDays:  ptr(30),
			componentName: []string{"alpha"},
			req:           GetRegistryPruneTagsRequest{GracePeriodOverride: ptr(time.Duration(0))},
			deploys: []pruneDeploySpec{
				{status: active, age: 5 * time.Hour},
				{status: errored, age: 20 * time.Minute},
				{status: active, age: 10 * time.Minute},
			},
			wantMode:     "enforce",
			wantGrace:    ptr(time.Duration(0)),
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{2},
			wantKeepIDs:  []int{2},
		},
		"zero grace period does not protect recent deploys": {
			appMode:       ptr("enforce"),
			appRetain:     ptr(1),
			componentName: []string{"alpha"},
			deploys: []pruneDeploySpec{
				{status: active, age: 5 * time.Hour},
				{status: errored, age: 20 * time.Minute},
				{status: active, age: 10 * time.Minute},
			},
			wantMode:     "enforce",
			wantOwned:    []int{0, 1, 2},
			wantKeepTags: []int{2},
			wantKeepIDs:  []int{2},
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.runCase(tc)
		})
	}
}

func (s *registryPruneTestSuite) runCase(tc pruneCase) {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	cfg := s.seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.seeder.CreateInstall(ctx, t, a)

	if tc.appMode != nil || tc.appRetain != nil || tc.appGraceDays != nil {
		require.NoError(t, s.db.WithContext(ctx).Create(&app.AppRegistryRetentionConfig{
			AppID:           a.ID,
			AppConfigID:     cfg.ID,
			Mode:            tc.appMode,
			RetainCount:     tc.appRetain,
			GracePeriodDays: tc.appGraceDays,
		}).Error)
	}
	if tc.installMode != nil {
		require.NoError(t, s.db.WithContext(ctx).Create(&app.InstallConfig{
			InstallID:             install.ID,
			RegistryRetentionMode: tc.installMode,
		}).Error)
	}

	require.LessOrEqual(t, len(tc.componentName), len(cfg.ComponentConfigConnections))
	installComponentIDs := make([]string, len(tc.componentName))
	cccIDs := make([]string, len(tc.componentName))
	for i, name := range tc.componentName {
		ccc := cfg.ComponentConfigConnections[i]
		require.NoError(t, s.db.WithContext(ctx).
			Model(&app.Component{}).
			Where(app.Component{ID: ccc.ComponentID}).
			Update("name", name).Error)
		if n, ok := tc.compRetain[i]; ok {
			require.NoError(t, s.db.WithContext(ctx).
				Model(&app.ComponentConfigConnection{}).
				Where(app.ComponentConfigConnection{ID: ccc.ID}).
				Update("registry_retain_count", n).Error)
		}
		ic := s.seeder.CreateInstallComponent(ctx, t, install.ID, ccc.ComponentID)
		installComponentIDs[i] = ic.ID
		cccIDs[i] = ccc.ID
	}

	now := time.Now()
	deployIDs := make([]string, len(tc.deploys))
	buildIDs := make([]string, len(tc.deploys))
	tags := make([]string, len(tc.deploys))
	for i, d := range tc.deploys {
		bld := s.seeder.CreateComponentBuild(ctx, t, cccIDs[d.component])
		if d.digest != "" || d.resolved != "" {
			require.NoError(t, s.db.WithContext(ctx).
				Model(&app.ComponentBuild{}).
				Where(app.ComponentBuild{ID: bld.ID}).
				Updates(map[string]any{"source_digest": d.digest, "resolved_tag": d.resolved}).Error)
		}
		dep := s.seeder.CreateInstallDeploy(ctx, t, installComponentIDs[d.component], bld.ID)
		require.NoError(t, s.db.WithContext(ctx).
			Model(&app.InstallDeploy{}).
			Where(app.InstallDeploy{ID: dep.ID}).
			Updates(map[string]any{"status": d.status, "created_at": now.Add(-d.age)}).Error)

		if d.status == app.InstallDeployStatusActive && !d.noPlan {
			plan := d.plan
			if plan == "" {
				plan = `{"noop":true}`
			}
			job := s.seeder.CreateRunnerJob(ctx, t, dep.ID, "install_deploys")
			require.NoError(t, s.db.WithContext(ctx).Create(&app.RunnerJobPlan{
				RunnerJobID: job.ID,
				PlanJSON:    plan,
			}).Error)
		}

		deployIDs[i] = dep.ID
		buildIDs[i] = bld.ID
		tags[i] = d.wantTag
		if tags[i] == "" {
			tags[i] = bld.ID
		}
	}

	acts := &Activities{
		db:          s.db,
		appsHelpers: appshelpers.New(appshelpers.Params{DB: s.db}),
	}
	req := tc.req
	req.InstallID = install.ID
	if req.GracePeriodOverride == nil && tc.appGraceDays == nil && tc.wantGrace == nil {
		req.GracePeriodOverride = ptr(time.Duration(0))
	}
	got, err := acts.GetRegistryPruneTags(ctx, &req)

	if tc.wantErr != "" {
		require.Error(t, err)
		require.Contains(t, err.Error(), tc.wantErr)
		return
	}
	require.NoError(t, err)
	require.Equal(t, tc.wantMode, got.Mode)
	if tc.wantGrace != nil {
		require.Equal(t, *tc.wantGrace, got.GracePeriod)
	}

	if tc.wantNoTags {
		require.Empty(t, got.OwnedTags)
		require.Empty(t, got.OwnedTagPatterns)
		require.Empty(t, got.KeepTags)
		return
	}

	require.Equal(t, []string{`^dpl[0-9a-z]{23}$`}, got.OwnedTagPatterns)
	require.Equal(t, sortedPick(tags, tc.wantOwned), got.OwnedTags)

	wantKeep := append(sortedPick(tags, tc.wantKeepTags), sortedPick(deployIDs, tc.wantKeepIDs)...)
	sort.Strings(wantKeep)
	require.Equal(t, wantKeep, got.KeepTags)
}

func sortedPick(values []string, idx []int) []string {
	out := make([]string, 0, len(idx))
	seen := map[string]bool{}
	for _, i := range idx {
		if seen[values[i]] {
			continue
		}
		seen[values[i]] = true
		out = append(out, values[i])
	}
	sort.Strings(out)
	return out
}
