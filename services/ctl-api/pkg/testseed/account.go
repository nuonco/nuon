package testseed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// EnsureAccount creates a test account and sets it in the context.
// Returns the updated context with the account ID set.
//
// The account is created with:
// - A unique fake subject ID
// - An email formatted as <subject>@test.nuon.co
// - Empty user journeys (no onboarding flow)
//
// The returned context has the account ID set via cctx.SetAccountIDContext,
// which is required for most ctl-api operations.
func (s *Seeder) EnsureAccount(ctx context.Context, t *testing.T) context.Context {
	subjectID := generics.GetFakeObj[string]()
	email := fmt.Sprintf("%s@test.nuon.co", subjectID)

	acct, err := s.acctHelpers.CreateAccount(ctx, email, subjectID, app.UserJourneys{})
	require.Nil(t, err)

	return cctx.SetAccountIDContext(ctx, acct.ID)
}
