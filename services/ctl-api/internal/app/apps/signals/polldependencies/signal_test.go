package polldependencies

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TODO: These tests are scaffolds and won't run until we have better seed tooling

type PollDependenciesSignalTestSuite struct {
	suite.Suite
}

func TestPollDependenciesSignalSuite(t *testing.T) {
	t.Skip("TODO: Tests need seed tooling - scaffold only")
	suite.Run(t, new(PollDependenciesSignalTestSuite))
}

func (s *PollDependenciesSignalTestSuite) SetupSuite() {
	// TODO: Initialize test dependencies with FX
}

func (s *PollDependenciesSignalTestSuite) TearDownSuite() {
	// TODO: Cleanup
}

func (s *PollDependenciesSignalTestSuite) TestPollCompletesWhenOrgActive() {
	// TODO: Test that polling completes when org becomes active
	require.True(s.T(), true, "placeholder test")
}

func (s *PollDependenciesSignalTestSuite) TestPollPropagatesOrgError() {
	// TODO: Test that org error status is propagated to app
	require.True(s.T(), true, "placeholder test")
}
