package syncstatus_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/syncstatus"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type readDeps struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	Seed     *testseed.Seeder
	Installs *installhelpers.Helpers
	Blob     blobstore.Service
	L        *zap.Logger
}

type ReadSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps readDeps
	ctx  context.Context
}

func TestReadSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(ReadSuite))
}

func (s *ReadSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:               s.T(),
		Mocks:           &tests.TestMocks{MockTC: temporalclient.NewMockClient(gomock.NewController(s.T()))},
		CustomValidator: true,
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *ReadSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *ReadSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
}

type readFixture struct {
	install *app.Install
	ic      *app.InstallComponent
	conn    *app.ComponentConfigConnection
	builds  map[string]string
	base    time.Time
}

func (s *ReadSuite) seed(buildNames ...string) *readFixture {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	cfg := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	comp := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	conn := s.deps.Seed.CreateTerraformComponentConfigConnection(s.ctx, s.T(), comp.ID, cfg.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(conn).Update("checksum", "checksum-1").Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Exec(
		"UPDATE app_configs SET component_ids = array_append(component_ids, ?) WHERE id = ?", comp.ID, cfg.ID).Error)
	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	f := &readFixture{
		install: install,
		ic:      s.deps.Seed.CreateInstallComponent(s.ctx, s.T(), install.ID, comp.ID),
		conn:    conn,
		builds:  map[string]string{},
		base:    time.Now().UTC().Add(-48 * time.Hour).Truncate(time.Second),
	}
	for _, name := range buildNames {
		s.addBuild(f, name)
	}
	return f
}

func (s *ReadSuite) addBuild(f *readFixture, name string) {
	source := name
	sum := sha256.Sum256([]byte(source))
	digest := hex.EncodeToString(sum[:])
	identity := sourceidentity.New("terraform", source)
	identity.Captured = true
	identity.Commit = digest[:40]
	identity.Repository = "https://example.com/acme/repo"
	identity.Files = []sourceidentity.File{{Path: "main.tf", SHA256: digest}}
	identity.Seal()
	build := app.ComponentBuild{
		ComponentConfigConnectionID: f.conn.ID,
		Status:                      app.ComponentBuildStatusActive,
		StatusDescription:           "active",
		SourceIdentity:              identity,
		SourceChecksum:              identity.Checksum,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&build).Error)
	f.builds[name] = build.ID
}

func (s *ReadSuite) target(f *readFixture, build string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(f.conn).
		Update("latest_build_id", f.builds[build]).Error)
}

type seededDeploy struct {
	build    string
	minute   int
	status   string
	applied  bool
	workflow *app.Workflow
}

func (s *ReadSuite) workflow(f *readFixture, status app.Status, planOnly bool) *app.Workflow {
	return s.deps.Seed.CreateWorkflow(s.ctx, s.T(), f.install.ID, app.WorkflowTypeDeployComponents,
		testseed.WithWorkflowStatus(app.CompositeStatus{Status: status}),
		testseed.WithWorkflowPlanOnly(planOnly),
		testseed.WithWorkflowCreatedAt(f.base))
}

func (s *ReadSuite) deploy(f *readFixture, d seededDeploy) *app.InstallDeploy {
	deploy := s.deps.Seed.CreateInstallDeploy(s.ctx, s.T(), f.ic.ID, f.builds[d.build])
	createdAt := f.base.Add(time.Duration(d.minute) * time.Minute)
	updates := map[string]any{
		"created_at": createdAt,
		"type":       app.InstallDeployTypeApply,
		"status":     d.status,
		"status_v2":  app.CompositeStatus{Status: app.Status(d.status)},
	}
	if d.applied {
		updates["applied_at"] = createdAt.Add(30 * time.Second)
	}
	if d.workflow != nil {
		updates["install_workflow_id"] = d.workflow.ID
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(deploy).Updates(updates).Error)
	return deploy
}

func (s *ReadSuite) read(f *readFixture) syncstatus.InstallResourceSyncStatusResource {
	var install app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&install, "id = ?", f.install.ID).Error)
	h := syncstatus.New(s.deps.DB, s.deps.Installs, s.deps.Blob, s.deps.L)
	got, err := h.Read(s.ctx, &install)
	require.NoError(s.T(), err)
	for _, res := range got.Resources {
		if res.ID == f.ic.ID {
			return res
		}
	}
	s.T().Fatalf("install component %s missing from sync status", f.ic.ID)
	return syncstatus.InstallResourceSyncStatusResource{}
}

func (s *ReadSuite) TestLiveBuildIsLatestApplied() {
	f := s.seed("A", "B", "C", "D", "E", "F", "G")
	var deployF *app.InstallDeploy
	for i, d := range []seededDeploy{
		{build: "A", status: "active", applied: true},
		{build: "B", status: "error"},
		{build: "C", status: "active", applied: true},
		{build: "D", status: "error"},
		{build: "E", status: "user-skipped"},
		{build: "F", status: "active", applied: true},
	} {
		d.minute = i + 1
		d.workflow = s.workflow(f, app.StatusSuccess, false)
		deployF = s.deploy(f, d)
	}

	s.target(f, "F")
	res := s.read(f)
	s.Equal(syncstatus.InstallResourceSyncStatusCategoryInSync, res.Category, res.Reason)
	s.Equal(f.builds["F"], res.AppliedBuildID)
	s.Equal(deployF.ID, res.AppliedDeployID)
	s.Require().NotNil(res.AppliedDeployAt)
	s.WithinDuration(f.base.Add(6*time.Minute+30*time.Second), *res.AppliedDeployAt, time.Millisecond)

	s.target(f, "G")
	s.deploy(f, seededDeploy{build: "G", minute: 7, status: "error", workflow: s.workflow(f, app.StatusError, false)})
	res = s.read(f)
	s.Equal(syncstatus.InstallResourceSyncStatusCategoryChangesNotApplied, res.Category)
	s.Equal(f.builds["G"], res.ExpectedBuildID)
	s.Equal(f.builds["F"], res.AppliedBuildID)
	s.Nil(res.OpenDeployment)
}

func (s *ReadSuite) TestOpenDeploymentComesFromLatestDeploy() {
	for name, tc := range map[string]struct {
		status   app.Status
		planOnly bool
		open     bool
	}{
		"in progress":            {status: app.StatusInProgress, open: true},
		"awaiting approval":      {status: app.AwaitingApproval, open: true},
		"failed awaiting retry":  {status: app.StatusFailedPendingRetry, open: true},
		"failed":                 {status: app.StatusError},
		"cancelled":              {status: app.StatusCancelled},
		"plan-only, in progress": {status: app.StatusInProgress, planOnly: true},
	} {
		s.Run(name, func() {
			f := s.seed("F", "G")
			s.deploy(f, seededDeploy{build: "F", minute: 1, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false)})
			s.target(f, "G")
			wf := s.workflow(f, tc.status, tc.planOnly)
			s.deploy(f, seededDeploy{build: "G", minute: 2, status: "queued", workflow: wf})

			res := s.read(f)
			s.Equal(syncstatus.InstallResourceSyncStatusCategoryChangesNotApplied, res.Category)
			if !tc.open {
				s.Nil(res.OpenDeployment)
				return
			}
			s.Require().NotNil(res.OpenDeployment)
			s.Equal(wf.ID, res.OpenDeployment.ID)
			s.Equal(string(tc.status), res.OpenDeployment.Status)
		})
	}
}
