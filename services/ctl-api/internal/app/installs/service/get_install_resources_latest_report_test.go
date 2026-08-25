package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type LatestReportTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

// Covers app.LatestReportOnlySQL, which is a raw predicate and so unreachable
// from a pure-Go test. Deletion is not representable in an append-only log: a
// removed resource stops being reported and the latest-state view keeps its
// final row forever, so a deleted pod read degraded permanently.
type LatestReportTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service LatestReportTestService

	orgID      string
	installID  string
	componentA string
	componentB string
	latest     time.Time
	previous   time.Time
}

func TestLatestReportSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LatestReportTestSuite))
}

func (s *LatestReportTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// Provided locally on purpose: the shared test options leave flowclient
		// out because its import tree would create a cycle.
		fx.Provide(flowclient.New),
		fx.Provide(New),
		fx.Populate(&s.service),
	)
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
	s.SetCHDB(s.service.CHDB)
}

func (s *LatestReportTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LatestReportTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	ctx := context.Background()
	ctx, _ = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, org := s.service.Seeder.EnsureOrg(ctx, s.T())
	testApp := s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.service.Seeder.CreateInstall(ctx, s.T(), testApp)

	// A fresh install per test: the ClickHouse table has a 7 day TTL and no
	// per-test truncation, so rows must not be able to leak between runs.
	s.orgID = org.ID
	s.installID = install.ID
	s.componentA = s.deployedComponent(ctx, testApp.ID, install.ID)
	s.componentB = s.deployedComponent(ctx, testApp.ID, install.ID)

	// Truncated to the second: observed_at is what groups a report, and a
	// sub-second skew would split one report into two.
	s.latest = time.Now().UTC().Truncate(time.Second)
	s.previous = s.latest.Add(-time.Minute)
}

// Resource rows are dropped for a component that never deployed, so the
// component has to look deployed for any of this to be reachable.
func (s *LatestReportTestSuite) deployedComponent(ctx context.Context, appID, installID string) string {
	comp := s.service.Seeder.CreateComponent(ctx, s.T(), appID, app.ComponentTypeHelmChart)
	ic := s.service.Seeder.CreateInstallComponent(ctx, s.T(), installID, comp.ID)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&app.InstallComponent{}).
		Where("id = ?", ic.ID).
		Update("status", app.InstallComponentStatusActive).Error)
	return ic.ID
}

func (s *LatestReportTestSuite) row(
	componentID, source, ownerName, provider, kind, name, health string,
	observedAt time.Time,
	staleAfter uint32,
) app.InstallComponentResourceState {
	return app.InstallComponentResourceState{
		OrgID:              s.orgID,
		InstallID:          s.installID,
		InstallComponentID: componentID,
		Source:             source,
		OwnerName:          ownerName,
		Provider:           provider,
		Kind:               kind,
		Name:               name,
		Health:             health,
		ObservedAt:         observedAt,
		StaleAfterSeconds:  staleAfter,
	}
}

func (s *LatestReportTestSuite) insert(rows []app.InstallComponentResourceState) {
	require.NoError(s.T(), s.service.CHDB.Create(&rows).Error)
}

func (s *LatestReportTestSuite) fetchNames() []string {
	got, err := s.service.InstallsService.getInstallResources(
		context.Background(), s.orgID, s.installID, installResourceFilters{})
	require.NoError(s.T(), err)

	names := make([]string, 0, len(got))
	for _, r := range got {
		names = append(names, r.Name)
	}
	return names
}

// The reported bug: a pod deleted by a rollout kept its final degraded row, and
// the dashboard read it as a live failure 14 minutes later.
func (s *LatestReportTestSuite) TestVanishedResourceIsDropped() {
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Pod", "old-pod", "degraded", s.previous, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Pod", "new-pod", "healthy", s.latest, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "api", "healthy", s.latest, 0),
	})

	names := s.fetchNames()

	assert.NotContains(s.T(), names, "old-pod", "a deleted resource must not be reported as live")
	assert.ElementsMatch(s.T(), []string{"new-pod", "api"}, names)
}

// Each component reports on its own cadence, so one component's newest report
// must not hide another's.
func (s *LatestReportTestSuite) TestScopedPerComponent() {
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "fresh", "healthy", s.latest, 0),
		s.row(s.componentB, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "lagging", "degraded", s.previous, 0),
	})

	names := s.fetchNames()

	assert.ElementsMatch(s.T(), []string{"fresh", "lagging"}, names,
		"a component's own newest report is what counts, not the install's")
}

// Sandbox rows carry no install_component_id and are keyed by owner_name, so
// grouping on the component alone would collapse every release into one.
func (s *LatestReportTestSuite) TestSandboxGroupedByOwner() {
	s.insert([]app.InstallComponentResourceState{
		s.row("", app.InstallComponentResourceSourceSandbox, "ingress-nginx", "kubernetes", "Deployment", "nginx-live", "healthy", s.latest, 0),
		s.row("", app.InstallComponentResourceSourceSandbox, "ingress-nginx", "kubernetes", "Pod", "nginx-gone", "degraded", s.previous, 0),
		s.row("", app.InstallComponentResourceSourceSandbox, "cert-manager", "kubernetes", "Deployment", "cert-live", "healthy", s.previous, 0),
	})

	names := s.fetchNames()

	assert.NotContains(s.T(), names, "nginx-gone", "a vanished sandbox resource is dropped within its release")
	assert.Contains(s.T(), names, "cert-live",
		"a different release reporting on its own cadence is not dropped by another's newer report")
	assert.Contains(s.T(), names, "nginx-live")
}

// Pushed checks arrive on their own cadence and expire by their own TTL, so the
// cluster report clock must not evict them.
func (s *LatestReportTestSuite) TestCustomChecksExempt() {
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "api", "healthy", s.latest, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", app.InstallComponentResourceProviderCustom, "Check", "vendor-migration", "healthy", s.previous, 1800),
	})

	names := s.fetchNames()

	assert.Contains(s.T(), names, "vendor-migration",
		"a pushed check older than the cluster report keeps its own window")
	assert.Contains(s.T(), names, "api")
}

// Terraform identity rows ride the same report as the workload rows.
func (s *LatestReportTestSuite) TestCloudIdentityRowsRideTheReport() {
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Deployment", "api", "healthy", s.latest, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "aws", "aws_acm_certificate", "cert-live", "not-applicable", s.latest, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "aws", "aws_route53_record", "record-gone", "not-applicable", s.previous, 0),
	})

	names := s.fetchNames()

	assert.Contains(s.T(), names, "cert-live")
	assert.NotContains(s.T(), names, "record-gone", "a cloud row absent from the newest report is gone too")
}

// The health filter is applied alongside the predicate, so the newest report has
// to be established independently of it — otherwise filtering to degraded makes
// a vanished degraded row the newest thing the query can see.
func (s *LatestReportTestSuite) TestFilterDoesNotResurrectVanishedRows() {
	s.insert([]app.InstallComponentResourceState{
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Pod", "old-bad-pod", "degraded", s.previous, 0),
		s.row(s.componentA, app.InstallComponentResourceSourceComponent, "", "kubernetes", "Pod", "new-pod", "healthy", s.latest, 0),
	})

	got, err := s.service.InstallsService.getInstallResources(
		context.Background(), s.orgID, s.installID, installResourceFilters{Health: "degraded"})
	require.NoError(s.T(), err)

	assert.Empty(s.T(), got,
		"filtering to degraded must not surface a row the newest report does not contain")
}
