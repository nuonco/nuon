// Integration tests: run with INTEGRATION=true against the migrated test database.
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type gcpRunnerAuthDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Seeder *testseed.Seeder
}

type GCPRunnerServiceAccountTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  gcpRunnerAuthDeps
}

func TestGCPRunnerServiceAccountSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(GCPRunnerServiceAccountTestSuite))
}

func (s *GCPRunnerServiceAccountTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *GCPRunnerServiceAccountTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *GCPRunnerServiceAccountTestSuite) installWithRunnerSA(ctx context.Context, a *app.App, email string) *app.Install {
	t := s.T()
	install := s.deps.Seeder.CreateInstall(ctx, t, a)
	require.NoError(t, s.deps.DB.WithContext(ctx).
		Model(&app.InstallStackOutputs{}).
		Where(app.InstallStackOutputs{InstallStackID: install.InstallStack.ID}).
		Update("data", dbgenerics.ToHstore(map[string]string{"runner_service_account_email": email})).Error)
	return install
}

// Identity-only GCP login trusts the runner service account to pick the
// install. A service account reused across installs must not authenticate,
// or one install's VM could claim another install's runner.
func (s *GCPRunnerServiceAccountTestSuite) TestEnsureGCPRunnerServiceAccountNotShared() {
	t := s.T()
	ctx := context.Background()
	ctx, _ = s.deps.Seeder.EnsureAccount(ctx, t)
	ctx, _ = s.deps.Seeder.EnsureOrg(ctx, t)
	a := s.deps.Seeder.CreateApp(ctx, t)
	s.deps.Seeder.CreateAppConfig(ctx, t, a.ID)

	svc := &service{db: s.deps.DB}

	unique := s.installWithRunnerSA(ctx, a, a.ID+"-unique-runner@acme-project.iam.gserviceaccount.com")
	require.NoError(t, svc.ensureGCPRunnerServiceAccountNotShared(ctx, unique.ID, a.ID+"-unique-runner@acme-project.iam.gserviceaccount.com"))

	shared := a.ID + "-shared-runner@acme-project.iam.gserviceaccount.com"
	first := s.installWithRunnerSA(ctx, a, shared)
	second := s.installWithRunnerSA(ctx, a, shared)
	require.Error(t, svc.ensureGCPRunnerServiceAccountNotShared(ctx, first.ID, shared))
	require.Error(t, svc.ensureGCPRunnerServiceAccountNotShared(ctx, second.ID, shared))

	require.NoError(t, s.deps.DB.WithContext(ctx).Delete(second).Error)
	require.NoError(t, svc.ensureGCPRunnerServiceAccountNotShared(ctx, first.ID, shared))
}
