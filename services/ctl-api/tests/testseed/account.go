package testseed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func BuildAccount() *app.Account {
	subjectID := generics.GetFakeObj[string]()
	return &app.Account{
		ID:           domains.NewAccountID(),
		Subject:      subjectID,
		Email:        fmt.Sprintf("%s@test.nuon.co", subjectID),
		AccountType:  app.AccountTypeAuth0,
		UserJourneys: app.UserJourneys{},
	}
}

func (s *Seeder) CreateAccount(ctx context.Context, t *testing.T) *app.Account {
	acct := BuildAccount()
	res := s.db.WithContext(ctx).Create(acct)
	require.NoError(t, res.Error)
	return acct
}

func (s *Seeder) CreateServiceAccount(ctx context.Context, t *testing.T, svcAcctID string) *app.Account {
	acct := BuildAccount()
	acct.Subject = svcAcctID
	acct.Email = account.ServiceAccountEmail(svcAcctID)
	acct.AccountType = app.AccountTypeService
	res := s.db.WithContext(ctx).Create(acct)
	require.NoError(t, res.Error)
	return acct
}

func (s *Seeder) EnsureAccount(ctx context.Context, t *testing.T) (context.Context, *app.Account) {
	acct := s.CreateAccount(ctx, t)
	return cctx.SetAccountIDContext(ctx, acct.ID), acct
}
