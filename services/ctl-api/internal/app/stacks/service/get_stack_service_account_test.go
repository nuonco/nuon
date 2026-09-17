// Integration tests: run with INTEGRATION=true against the migrated test database.
package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type stackTokenExpiryDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Seeder *testseed.Seeder
}

type StackTokenExpiryTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  stackTokenExpiryDeps
}

func TestStackTokenExpirySuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(StackTokenExpiryTestSuite))
}

func (s *StackTokenExpiryTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *StackTokenExpiryTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

// revoke soft-deletes a token the way production revocation does.
func (s *StackTokenExpiryTestSuite) revoke(tok *app.Token) {
	require.NoError(s.T(), s.deps.DB.Delete(tok).Error)
}

// liveStackTokenExpiry drives what the TF Module tab tells the customer: whether they
// still hold a working credential, and when it dies. Getting it wrong either claims a
// stack is authenticated when it is not, or prompts a duplicate token.
//
// Each case seeds its own account, so the shared database isolates by account ID.
func (s *StackTokenExpiryTestSuite) TestLiveStackTokenExpiry() {
	t := s.T()
	ctx := context.Background()
	future := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	farFuture := time.Now().Add(365 * 24 * time.Hour).UTC().Truncate(time.Second)
	past := time.Now().Add(-24 * time.Hour)

	s.Run("no token at all", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "no token means no expiry to report")
	})

	s.Run("live token reports its expiry", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		s.deps.Seeder.CreateToken(ctx, t, acct, future)

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, future, expiry, time.Second)
	})

	s.Run("expired token", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		s.deps.Seeder.CreateToken(ctx, t, acct, past)

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "an expired token is not a credential")
	})

	s.Run("revoked token", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		s.revoke(s.deps.Seeder.CreateToken(ctx, t, acct, future))

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "a soft-deleted token must be filtered out by gorm")
	})

	s.Run("another account's token does not count", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		other := s.deps.Seeder.CreateAccount(ctx, t)
		s.deps.Seeder.CreateToken(ctx, t, other, future)

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.True(t, expiry.IsZero())
	})

	// Ordered by expiry, not created_at: the newest token can be the shortest-lived.
	s.Run("reports the longest-lived token, not the newest", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		s.deps.Seeder.CreateToken(ctx, t, acct, farFuture)
		s.deps.Seeder.CreateToken(ctx, t, acct, future)

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, farFuture, expiry, time.Second)
	})

	s.Run("expired and revoked tokens together report nothing live", func() {
		acct := s.deps.Seeder.CreateAccount(ctx, t)
		s.deps.Seeder.CreateToken(ctx, t, acct, past)
		s.revoke(s.deps.Seeder.CreateToken(ctx, t, acct, future))

		expiry, err := liveStackTokenExpiry(ctx, s.deps.DB, acct.ID)
		require.NoError(t, err)
		assert.True(t, expiry.IsZero())
	})
}

type stackRunnerAPIURLDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Cfg    *internal.Config
	Seeder *testseed.Seeder
}

type StackRunnerAPIURLTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  stackRunnerAPIURLDeps
}

func TestStackRunnerAPIURLSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(StackRunnerAPIURLTestSuite))
}

func (s *StackRunnerAPIURLTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *StackRunnerAPIURLTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *StackRunnerAPIURLTestSuite) seedInstall(ctx context.Context) (context.Context, *app.Install) {
	t := s.T()
	ctx, _ = s.deps.Seeder.EnsureAccount(ctx, t)
	ctx, _ = s.deps.Seeder.EnsureOrg(ctx, t)
	a := s.deps.Seeder.CreateApp(ctx, t)
	s.deps.Seeder.CreateAppConfig(ctx, t, a.ID)

	return ctx, s.deps.Seeder.CreateInstall(ctx, t, a)
}

// A wrong answer points the customer's Terraform at the wrong control plane.
func (s *StackRunnerAPIURLTestSuite) TestStackRunnerAPIURL() {
	t := s.T()
	const globalURL = "https://runner.example.com"

	svc := &service{db: s.deps.DB, cfg: &internal.Config{RunnerAPIURL: globalURL}}

	s.Run("falls back to the global config when the group has no setting", func() {
		ctx, install := s.seedInstall(context.Background())

		url, err := svc.stackRunnerAPIURL(ctx, install.ID, install.OrgID)
		require.NoError(t, err)
		assert.Equal(t, globalURL, url)
	})

	s.Run("the runner group setting wins", func() {
		ctx, install := s.seedInstall(context.Background())
		const groupURL = "https://runner.byoc.example.net"

		group := &app.RunnerGroup{
			ID:        domains.NewRunnerGroupID(),
			OrgID:     install.OrgID,
			OwnerID:   install.ID,
			OwnerType: "installs",
			Type:      app.RunnerGroupTypeInstall,
			Settings:  app.RunnerGroupSettings{RunnerAPIURL: groupURL},
		}
		require.NoError(t, s.deps.DB.WithContext(ctx).Create(group).Error)

		url, err := svc.stackRunnerAPIURL(ctx, install.ID, install.OrgID)
		require.NoError(t, err)
		assert.Equal(t, groupURL, url)
	})

	// The endpoint answers not-found for it, so this lookup must not error first.
	s.Run("an unknown install still reports the global config", func() {
		ctx, install := s.seedInstall(context.Background())

		url, err := svc.stackRunnerAPIURL(ctx, "inl00000000000000000000000", install.OrgID)
		require.NoError(t, err)
		assert.Equal(t, globalURL, url)
	})
}
