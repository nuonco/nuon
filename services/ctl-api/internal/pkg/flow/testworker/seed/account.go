package seed

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// fakeMu serializes go-faker access: faker mutates package-global state and
// the seeder runs from concurrent test cases.
var fakeMu sync.Mutex

func FakeString() string {
	fakeMu.Lock()
	defer fakeMu.Unlock()
	return generics.GetFakeObj[string]()
}

func (s *Seeder) EnsureAccount(ctx context.Context, t *testing.T) context.Context {
	subjectID := FakeString()
	email := fmt.Sprintf("%s@test.nuon.co", subjectID)

	acct, err := s.AcctHelpers.CreateAccount(ctx, email, subjectID, app.UserJourneys{})
	require.Nil(t, err)

	return cctx.SetAccountIDContext(ctx, acct.ID)
}
