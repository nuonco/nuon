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

// revoke soft-deletes a token the way production revocation does, rather than
// inserting a pre-cooked deleted row.
func (s *StackTokenExpiryTestSuite) revoke(tok *app.Token) {
	require.NoError(s.T(), s.deps.DB.Delete(tok).Error)
}

// liveStackTokenExpiry drives what the TF Module tab tells the customer: whether they
// still have a working credential, and when it dies. A wrong answer here either tells
// someone their stack is authenticated when it is not, or prompts them to mint a
// duplicate token they did not need.
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

	// The reason this orders by expiry and not by created_at. The create modal lets
	// the caller pick a duration, so the newest token can be the shortest-lived one;
	// reporting its expiry would warn a customer holding a year-long credential that
	// they expire tomorrow.
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
