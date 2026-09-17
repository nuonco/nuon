package activities_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type hydrateLogStreamDeps struct {
	fx.In

	DB         *gorm.DB `name:"psql"`
	Cfg        *internal.Config
	AcctClient *account.Client
	Activities *activities.Activities
	Seeder     *testseed.Seeder
}

type HydrateLogStreamTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  hydrateLogStreamDeps
}

func TestHydrateLogStreamSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(HydrateLogStreamTestSuite))
}

func (s *HydrateLogStreamTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(activities.New),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *HydrateLogStreamTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *HydrateLogStreamTestSuite) TestHydratesCredentialsWithoutPersistingThemOnLogStream() {
	t := s.T()
	ctx, _ := s.deps.Seeder.EnsureAccount(context.Background(), t)
	ctx, org := s.deps.Seeder.EnsureOrg(ctx, t)

	stream := &app.LogStream{OwnerType: "test", Open: true}
	require.NoError(t, s.deps.DB.WithContext(ctx).Create(stream).Error)
	_, err := s.deps.AcctClient.CreateServiceAccount(ctx, stream.ID, "")
	require.NoError(t, err)

	originalRunnerAPIURL := s.deps.Cfg.RunnerAPIURL
	s.deps.Cfg.RunnerAPIURL = "https://runner.example.com"
	t.Cleanup(func() {
		s.deps.Cfg.RunnerAPIURL = originalRunnerAPIURL
	})

	hydrated, err := s.deps.Activities.HydrateLogStream(ctx, &activities.HydrateLogStreamRequest{
		LogStreamID: stream.ID,
		OrgID:       org.ID,
	})
	require.NoError(t, err)
	require.Equal(t, stream.ID, hydrated.ID)
	require.Equal(t, "https://runner.example.com", hydrated.RunnerAPIURL)
	require.NotEmpty(t, hydrated.WriteToken)

	var token app.Token
	require.NoError(t, s.deps.DB.WithContext(ctx).
		Where(app.Token{Token: hydrated.WriteToken}).
		First(&token).Error)
	require.True(t, token.ExpiresAt.After(time.Now().Add(55*time.Minute)))

	var persisted app.LogStream
	require.NoError(t, s.deps.DB.WithContext(ctx).First(&persisted, app.LogStream{ID: stream.ID}).Error)
	require.Empty(t, persisted.RunnerAPIURL)
	require.Empty(t, persisted.WriteToken)
}

func (s *HydrateLogStreamTestSuite) TestRejectsAnotherOrgLogStream() {
	t := s.T()
	ctx, _ := s.deps.Seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.deps.Seeder.EnsureOrg(ctx, t)

	stream := &app.LogStream{OwnerType: "test", Open: true}
	require.NoError(t, s.deps.DB.WithContext(ctx).Create(stream).Error)

	_, err := s.deps.Activities.HydrateLogStream(ctx, &activities.HydrateLogStreamRequest{
		LogStreamID: stream.ID,
		OrgID:       "org00000000000000000000000",
	})
	require.Error(t, err)
}
