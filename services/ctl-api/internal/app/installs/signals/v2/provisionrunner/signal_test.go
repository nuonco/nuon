package provisionrunner

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TODO: These tests are scaffolds and won't run until we have better seed tooling

type ProvisionRunnerSignalTestSuite struct {
	suite.Suite
}

func TestProvisionRunnerSignalSuite(t *testing.T) {
	t.Skip("TODO: Tests need seed tooling - scaffold only")
	suite.Run(t, new(ProvisionRunnerSignalTestSuite))
}

func (s *ProvisionRunnerSignalTestSuite) SetupSuite() {
	// TODO: Initialize test dependencies with FX
}

func (s *ProvisionRunnerSignalTestSuite) TearDownSuite() {
	// TODO: Cleanup test dependencies
}

func (s *ProvisionRunnerSignalTestSuite) TestProvisionRunnerSignalExecutesSuccessfully() {
	// TODO: Implement test when seed tooling is ready
	// Steps:
	// 1. Create test install with runner
	// 2. Enqueue provision-runner signal
	// 3. Verify signal was sent to runner's event loop/queue
	require.True(s.T(), true, "placeholder test")
}

func (s *ProvisionRunnerSignalTestSuite) TestProvisionRunnerSignalValidationFails() {
	// TODO: Test validation with empty InstallID
	require.True(s.T(), true, "placeholder test")
}
