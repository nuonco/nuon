package alignment_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type attemptsDeps struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	Seed     *testseed.Seeder
	Installs *installhelpers.Helpers
	Blob     blobstore.Service
	L        *zap.Logger
}

type BlockingAttemptsSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps attemptsDeps
	ctx  context.Context
}

func TestBlockingAttemptsSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(BlockingAttemptsSuite))
}

func (s *BlockingAttemptsSuite) SetupSuite() {
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

func (s *BlockingAttemptsSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *BlockingAttemptsSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
}

type attemptFixture struct {
	install *app.Install
	ic      *app.InstallComponent
	conn    *app.ComponentConfigConnection
	builds  map[string]string
	base    time.Time
}

func (s *BlockingAttemptsSuite) seed(buildNames ...string) *attemptFixture {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	cfg := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	comp := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	conn := s.deps.Seed.CreateTerraformComponentConfigConnection(s.ctx, s.T(), comp.ID, cfg.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(conn).Update("checksum", "checksum-1").Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Exec(
		"UPDATE app_configs SET component_ids = array_append(component_ids, ?) WHERE id = ?", comp.ID, cfg.ID).Error)
	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	f := &attemptFixture{
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

func (s *BlockingAttemptsSuite) addBuild(f *attemptFixture, name string) {
	s.addBuildWithSource(f, name, name)
}

func (s *BlockingAttemptsSuite) addBuildWithSource(f *attemptFixture, name, source string) {
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

func (s *BlockingAttemptsSuite) target(f *attemptFixture, build string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(f.conn).
		Update("latest_build_id", f.builds[build]).Error)
}

type seededDeploy struct {
	build      string
	minute     int
	typ        app.InstallDeployType
	status     string
	v2Metadata map[string]any
	applied    bool
	workflow   *app.Workflow
	steps      []app.CompositeStatus
	retried    []bool
}

func (s *BlockingAttemptsSuite) workflow(f *attemptFixture, status app.Status, planOnly bool) *app.Workflow {
	return s.deps.Seed.CreateWorkflow(s.ctx, s.T(), f.install.ID, app.WorkflowTypeDeployComponents,
		testseed.WithWorkflowStatus(app.CompositeStatus{Status: status}),
		testseed.WithWorkflowPlanOnly(planOnly),
		testseed.WithWorkflowCreatedAt(f.base))
}

func (s *BlockingAttemptsSuite) deploy(f *attemptFixture, d seededDeploy) (*app.InstallDeploy, []*app.WorkflowStep) {
	deploy := s.deps.Seed.CreateInstallDeploy(s.ctx, s.T(), f.ic.ID, f.builds[d.build])
	createdAt := f.base.Add(time.Duration(d.minute) * time.Minute)
	typ := d.typ
	if typ == "" {
		typ = app.InstallDeployTypeApply
	}
	updates := map[string]any{
		"created_at": createdAt,
		"type":       typ,
		"status":     d.status,
		"status_v2":  app.CompositeStatus{Status: app.Status(d.status), Metadata: d.v2Metadata},
	}
	if d.applied {
		updates["applied_at"] = createdAt.Add(30 * time.Second)
	}
	if d.workflow != nil {
		updates["install_workflow_id"] = d.workflow.ID
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(deploy).Updates(updates).Error)

	var steps []*app.WorkflowStep
	for i, status := range d.steps {
		retried := i < len(d.retried) && d.retried[i]
		steps = append(steps, s.deps.Seed.CreateWorkflowStep(s.ctx, s.T(), d.workflow.ID,
			testseed.WithStepStatus(status),
			func(step *app.WorkflowStep) {
				step.OwnerID = f.install.ID
				step.OwnerType = "installs"
				step.StepTargetID = deploy.ID
				step.StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
				step.Idx = i
				step.Retried = retried
			}))
	}
	return deploy, steps
}

func (s *BlockingAttemptsSuite) read(f *attemptFixture) alignment.InstallAlignmentResource {
	var install app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&install, "id = ?", f.install.ID).Error)
	h := alignment.New(s.deps.DB, s.deps.Installs, s.deps.Blob, s.deps.L)
	got, err := h.Read(s.ctx, &install)
	require.NoError(s.T(), err)
	for _, res := range got.Resources {
		if res.ID == f.ic.ID {
			return res
		}
	}
	s.T().Fatalf("install component %s missing from alignment", f.ic.ID)
	return alignment.InstallAlignmentResource{}
}

func denied(status app.Status) app.CompositeStatus {
	return app.CompositeStatus{Status: status, Metadata: map[string]any{
		"deny_violations": []any{map[string]any{"policy_id": "pol-1", "message": "public buckets are not allowed"}},
	}}
}

func (s *BlockingAttemptsSuite) applyHistory(f *attemptFixture) *app.InstallDeploy {
	success := []app.CompositeStatus{{Status: app.WorkflowStepApprovalStatusApproved}, {Status: app.StatusSuccess}}
	s.deploy(f, seededDeploy{build: "A", minute: 1, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false), steps: success})
	s.deploy(f, seededDeploy{build: "B", minute: 2, status: "error", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}})
	s.deploy(f, seededDeploy{build: "C", minute: 3, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false), steps: success})
	s.deploy(f, seededDeploy{build: "D", minute: 4, status: "pending-approval", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{denied(app.StatusError)}})
	s.deploy(f, seededDeploy{build: "E", minute: 5, status: "error", workflow: s.workflow(f, app.StatusSuccess, false),
		steps: []app.CompositeStatus{{Status: app.StatusUserSkipped}}})
	deployF, _ := s.deploy(f, seededDeploy{build: "F", minute: 6, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false), steps: success})
	return deployF
}

func (s *BlockingAttemptsSuite) TestHistoryBeforeLastApplyIsNotShown() {
	f := s.seed("A", "B", "C", "D", "E", "F")
	deployF := s.applyHistory(f)

	s.target(f, "F")
	res := s.read(f)
	s.Equal(alignment.InstallAlignmentCategoryInSync, res.Category, res.Reason)
	s.Nil(res.BlockingAttempt)
	s.Zero(res.EarlierAttempts)
	s.Nil(res.Explanation)
	s.Require().NotNil(res.AppliedBuild)
	s.Equal(f.builds["F"], res.AppliedBuild.ID)
	s.Equal(deployF.ID, res.AppliedDeployID)
	s.Require().NotNil(res.AppliedDeployAt)
	s.WithinDuration(f.base.Add(6*time.Minute+30*time.Second), *res.AppliedDeployAt, time.Millisecond)

	s.addBuild(f, "G")
	s.target(f, "G")
	res = s.read(f)
	s.Equal(alignment.InstallAlignmentCategoryChangesNotApplied, res.Category)
	s.Nil(res.BlockingAttempt)
	s.Zero(res.EarlierAttempts)
	s.Require().NotNil(res.Explanation)
	s.Equal("No deployment has started for these changes.", res.Explanation.Summary)
	s.Empty(res.Explanation.WorkflowID)
	s.Require().NotNil(res.ExpectedBuild)
	s.Equal(f.builds["G"], res.ExpectedBuild.ID)
	s.Equal(f.builds["F"], res.AppliedBuild.ID)
}

func (s *BlockingAttemptsSuite) TestNewestRealAttemptAtWantedBuildBlocks() {
	f := s.seed("A", "B", "C", "D", "E", "F", "G")
	s.applyHistory(f)
	s.target(f, "G")

	s.deploy(f, seededDeploy{build: "G", minute: 0, status: "error", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}})
	s.deploy(f, seededDeploy{build: "D", minute: 7, status: "error", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}})
	s.deploy(f, seededDeploy{build: "G", minute: 8, status: "pending-approval", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{denied(app.StatusError)}})
	skippedWF := s.workflow(f, app.StatusCancelled, false)
	_, skippedSteps := s.deploy(f, seededDeploy{build: "G", minute: 9, status: "error", workflow: skippedWF,
		steps: []app.CompositeStatus{{Status: app.StatusUserSkipped}}})
	s.deploy(f, seededDeploy{build: "G", minute: 10, status: "queued"})
	s.deploy(f, seededDeploy{build: "G", minute: 11, status: "retried", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}})
	s.deploy(f, seededDeploy{build: "G", minute: 12, status: "drifted", workflow: s.workflow(f, app.StatusSuccess, true),
		steps: []app.CompositeStatus{{Status: app.WorkflowStepApprovalStatusApproved}}})
	s.deploy(f, seededDeploy{build: "G", minute: 13, status: "error", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}, retried: []bool{true}})

	res := s.read(f)
	s.Require().NotNil(res.BlockingAttempt)
	s.Equal("user-skipped", res.BlockingAttempt.Status)
	s.Equal(skippedWF.ID, res.BlockingAttempt.WorkflowID)
	s.Equal(skippedSteps[0].ID, res.BlockingAttempt.StepID)
	s.Equal(f.builds["G"], res.BlockingAttempt.BuildID)
	s.Equal(1, res.EarlierAttempts)
	s.Require().NotNil(res.Explanation)
	s.Equal("user-skipped", res.Explanation.Status)
	s.Equal(skippedWF.ID, res.Explanation.WorkflowID)
	s.Equal("The deployment of this resource was skipped.", res.Explanation.Summary)
}

func (s *BlockingAttemptsSuite) TestOutcomeComesFromStepsNotDeployRow() {
	for name, tc := range map[string]struct {
		rowStatus  string
		v2Metadata map[string]any
		wfStatus   app.Status
		noWorkflow bool
		steps      []app.CompositeStatus
		retried    []bool
		want       string
		summary    string
	}{
		"policy deny leaves the row at pending-approval": {
			rowStatus: "pending-approval", wfStatus: app.StatusError,
			steps: []app.CompositeStatus{denied(app.StatusError)},
			want:  "policy-failed", summary: "A policy denied the latest deployment.",
		},
		"user skip after a policy deny": {
			rowStatus: "pending-approval", wfStatus: app.StatusError,
			steps: []app.CompositeStatus{denied(app.StatusUserSkipped)},
			want:  "user-skipped", summary: "A policy denied the latest deployment, and it was skipped.",
		},
		"empty row status, workflow stopped before the step": {
			rowStatus: "", wfStatus: app.StatusError,
			want: "not-attempted", summary: "The deployment stopped before it reached this resource.",
		},
		"empty row status, step never started, workflow failed elsewhere": {
			rowStatus: "", wfStatus: app.StatusError,
			steps: []app.CompositeStatus{{Status: ""}},
			want:  "not-attempted",
		},
		"denied approval recorded only on the row": {
			rowStatus: "approval-denied", wfStatus: app.StatusError,
			steps: []app.CompositeStatus{{Status: app.WorkflowStepApprovalStatusApproved}},
			want:  "approval-denied",
		},
		"approval denied on the step": {
			rowStatus: "pending-approval", wfStatus: app.StatusError,
			steps: []app.CompositeStatus{{Status: app.WorkflowStepApprovalStatusApprovalDenied}},
			want:  "approval-denied",
		},
		"policy skip_apply": {
			rowStatus: "auto-skipped", v2Metadata: map[string]any{"policy_action": "skip_apply"}, wfStatus: app.StatusSuccess,
			steps: []app.CompositeStatus{{Status: app.StatusAutoSkipped, Metadata: map[string]any{"check": "policy-skip-apply"}}},
			want:  "policy-skipped",
		},
		"cancelled workflow with the plan still running": {
			rowStatus: "planning", wfStatus: app.StatusCancelled,
			steps: []app.CompositeStatus{{Status: app.StatusPlanning}},
			want:  "cancelled", summary: "The deployment was cancelled before these changes took effect.",
		},
		"failed deploy outside any workflow": {
			rowStatus: "error", noWorkflow: true,
		},
		"orphaned queued row": {
			rowStatus: "queued", noWorkflow: true,
		},
		"orphaned empty row": {
			rowStatus: "", noWorkflow: true,
		},
		"step-less row in a successful workflow": {
			rowStatus: "queued", wfStatus: app.StatusSuccess,
		},
		"only retried steps": {
			rowStatus: "error", wfStatus: app.StatusError,
			steps:   []app.CompositeStatus{{Status: app.StatusError}},
			retried: []bool{true},
		},
	} {
		s.Run(name, func() {
			f := s.seed("F", "G")
			s.deploy(f, seededDeploy{build: "F", minute: 1, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false),
				steps: []app.CompositeStatus{{Status: app.StatusSuccess}}})
			s.target(f, "G")
			d := seededDeploy{build: "G", minute: 2, status: tc.rowStatus, v2Metadata: tc.v2Metadata, steps: tc.steps, retried: tc.retried}
			if !tc.noWorkflow {
				d.workflow = s.workflow(f, tc.wfStatus, false)
			}
			s.deploy(f, d)

			res := s.read(f)
			s.Equal(alignment.InstallAlignmentCategoryChangesNotApplied, res.Category)
			if tc.want == "" {
				s.Nil(res.BlockingAttempt)
				s.Require().NotNil(res.Explanation)
				s.True(strings.HasPrefix(res.Explanation.Summary, "No deployment has started"), res.Explanation.Summary)
				return
			}
			s.Require().NotNil(res.BlockingAttempt)
			s.Equal(tc.want, res.BlockingAttempt.Status)
			s.Require().NotNil(res.Explanation)
			s.Equal(tc.want, res.Explanation.Status)
			if tc.summary != "" {
				s.Equal(tc.summary, res.Explanation.Summary)
			}
		})
	}
}

func (s *BlockingAttemptsSuite) TestAttemptAtEquivalentBuildBlocks() {
	f := s.seed("F", "G")
	s.addBuildWithSource(f, "G-rebuild", "G")
	s.deploy(f, seededDeploy{build: "F", minute: 1, status: "active", applied: true, workflow: s.workflow(f, app.StatusSuccess, false),
		steps: []app.CompositeStatus{{Status: app.StatusSuccess}}})
	s.target(f, "G")
	s.deploy(f, seededDeploy{build: "G-rebuild", minute: 2, status: "error", workflow: s.workflow(f, app.StatusError, false),
		steps: []app.CompositeStatus{{Status: app.StatusError}}})

	res := s.read(f)
	s.Require().NotNil(res.BlockingAttempt)
	s.Equal("error", res.BlockingAttempt.Status)
	s.Equal(f.builds["G-rebuild"], res.BlockingAttempt.BuildID)
	s.Equal("The latest deployment failed.", res.Explanation.Summary)
}
