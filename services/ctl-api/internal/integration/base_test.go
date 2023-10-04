package integration

import (
	"context"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type baseIntegrationTestSuite struct {
	suite.Suite

	v            *validator.Validate
	apiClient    nuon.Client
	intAPIClient api.Client
	ctx          context.Context
	ctxCancel    func()
}

func (s *baseIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	s.ctx = ctx
	s.ctxCancel = ctxCancel

	s.v = validator.New()

	apiURL := os.Getenv("INTEGRATION_API_URL")
	assert.NotEmpty(s.T(), apiURL)

	apiToken := os.Getenv("INTEGRATION_API_TOKEN")
	assert.NotEmpty(s.T(), apiToken)

	apiClient, err := nuon.New(s.v,
		nuon.WithAuthToken(apiToken),
		nuon.WithURL(apiURL),
	)
	assert.NoError(s.T(), err)
	s.apiClient = apiClient

	internalAPIURL := os.Getenv("INTEGRATION_INTERNAL_API_URL")
	assert.NotEmpty(s.T(), internalAPIURL)

	intApiClient, err := api.New(s.v,
		api.WithURL(internalAPIURL),
	)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), intApiClient)
	s.intAPIClient = intApiClient
}

func (s *baseIntegrationTestSuite) deleteOrg(orgID string) {
	disabled := os.Getenv("INTEGRATION_NO_CLEANUP")
	if disabled != "" {
		return
	}

	err := s.intAPIClient.DeleteOrg(s.ctx, orgID)
	require.NoError(s.T(), err)
}
