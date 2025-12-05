package seed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (s *Seeder) EnsureAccount(ctx context.Context, t *testing.T) context.Context {
	subjectID := generics.GetFakeObj[string]()
	email := fmt.Sprintf("%s@test.nuon.co", subjectID)

	acct, err := s.acctHelpers.CreateAccount(ctx, email, subjectID)
	require.Nil(t, err)

	return cctx.SetAccountIDContext(ctx, acct.ID)
}
