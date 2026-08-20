package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type AccountIdentityTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type AccountIdentityTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service AccountIdentityTestService
}

func TestAccountIdentitySuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AccountIdentityTestSuite))
}

func (s *AccountIdentityTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AccountIdentityTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AccountIdentityTestSuite) seedAccount(ctx context.Context, email string) *app.Account {
	acct := &app.Account{
		ID:          domains.NewAccountID(),
		Subject:     domains.NewAccountID(),
		Email:       email,
		AccountType: app.AccountTypeAuth,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(acct).Error)
	return acct
}

func (s *AccountIdentityTestSuite) seedOIDCProvider(ctx context.Context, name, issuerURL string) *app.IdentityProvider {
	ip := &app.IdentityProvider{
		ID:      domains.NewIdentityProviderID(),
		Enabled: true,
	}
	require.NoError(s.T(), ip.SetOpenIDConfig(&providers.OpenIDConfig{
		BaseConfig: providers.BaseConfig{
			ClientID:     name + "-client-id",
			ClientSecret: name + "-client-secret",
			RedirectURL:  s.service.Cfg.NuonAuthRedirectURL,
		},
		IssuerURL: issuerURL,
	}))
	ip.Name = name
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(ip).Error)
	return ip
}

func (s *AccountIdentityTestSuite) identitiesFor(ctx context.Context, accountID string) []app.AccountIdentity {
	var identities []app.AccountIdentity
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Where(&app.AccountIdentity{AccountID: accountID}).
		Find(&identities).Error)
	return identities
}

// Before identities were keyed on identity_provider_id, a user with an existing OIDC identity who
// signed in through a second OIDC provider hit the (account_id, provider_type) unique index and got
// a 500 at login.
func (s *AccountIdentityTestSuite) TestOneAccountCanHoldTwoIdentitiesOfTheSameProviderType() {
	ctx := context.Background()

	email := fmt.Sprintf("%s@test.nuon.co", domains.NewAccountID())
	acct := s.seedAccount(ctx, email)

	first := s.seedOIDCProvider(ctx, "first", "https://first.example.com")
	second := s.seedOIDCProvider(ctx, "second", "https://second.example.com")

	_, err := s.service.AuthService.linkIdentityToAccount(ctx, acct, first, &providers.UserInfo{
		Subject: "first|subject",
		Email:   email,
	})
	require.NoError(s.T(), err)

	_, err = s.service.AuthService.linkIdentityToAccount(ctx, acct, second, &providers.UserInfo{
		Subject: "second|subject",
		Email:   email,
	})
	require.NoError(s.T(), err)

	identities := s.identitiesFor(ctx, acct.ID)
	require.Len(s.T(), identities, 2)

	providerIDs := map[string]bool{}
	for _, identity := range identities {
		providerIDs[identity.IdentityProviderID] = true
		require.Equal(s.T(), app.ProviderTypeOIDC, identity.ProviderType)
	}
	require.True(s.T(), providerIDs[first.ID])
	require.True(s.T(), providerIDs[second.ID])
}

// Two providers of one type can mint the same sub. Keyed on provider_type alone, the second one
// would have authenticated into the first one's account.
func (s *AccountIdentityTestSuite) TestSameSubFromTwoProvidersResolvesToDifferentAccounts() {
	ctx := context.Background()

	firstEmail := fmt.Sprintf("%s@test.nuon.co", domains.NewAccountID())
	secondEmail := fmt.Sprintf("%s@test.nuon.co", domains.NewAccountID())
	firstAcct := s.seedAccount(ctx, firstEmail)
	secondAcct := s.seedAccount(ctx, secondEmail)

	first := s.seedOIDCProvider(ctx, "collide-first", "https://first.example.com")
	second := s.seedOIDCProvider(ctx, "collide-second", "https://second.example.com")

	const sharedSub = "shared|subject"

	_, err := s.service.AuthService.linkIdentityToAccount(ctx, firstAcct, first, &providers.UserInfo{
		Subject: sharedSub,
		Email:   firstEmail,
	})
	require.NoError(s.T(), err)

	_, err = s.service.AuthService.linkIdentityToAccount(ctx, secondAcct, second, &providers.UserInfo{
		Subject: sharedSub,
		Email:   secondEmail,
	})
	require.NoError(s.T(), err)

	resolved, err := s.service.AuthService.getOrCreateAccountByIdentityStrict(ctx, second, &providers.UserInfo{
		Subject: sharedSub,
		Email:   secondEmail,
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), secondAcct.ID, resolved.ID)
}
