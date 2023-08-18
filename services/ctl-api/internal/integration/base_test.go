package integration

import (
	"context"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type baseIntegrationTestSuite struct {
	suite.Suite

	v         *validator.Validate
	apiClient client.Client
	ctx       context.Context
	ctxCancel func()
}

func (s *baseIntegrationTestSuite) SetupTest() {
	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	s.ctx = ctx
	s.ctxCancel = ctxCancel

	s.v = validator.New()

	apiURL := os.Getenv("INTEGRATION_API_URL")
	assert.NotEmpty(s.T(), apiURL)

	apiToken := os.Getenv("INTEGRATION_API_TOKEN")
	assert.NotEmpty(s.T(), apiToken)

	apiClient, err := client.New(s.v,
		client.WithAuthToken(apiToken),
		client.WithURL(apiURL),
	)
	assert.NoError(s.T(), err)
	s.apiClient = apiClient
}
