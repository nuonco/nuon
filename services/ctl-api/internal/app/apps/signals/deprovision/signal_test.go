package deprovision

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TODO: These tests are scaffolds and won't run until we have better seed tooling

type DeprovisionSignalTestSuite struct {
	suite.Suite
}

func TestDeprovisionSignalSuite(t *testing.T) {
	t.Skip("TODO: Tests need seed tooling - scaffold only")
	suite.Run(t, new(DeprovisionSignalTestSuite))
}

func (s *DeprovisionSignalTestSuite) SetupSuite() {
	// TODO: Initialize test dependencies
}

func (s *DeprovisionSignalTestSuite) TearDownSuite() {
	// TODO: Cleanup
}

func (s *DeprovisionSignalTestSuite) TestDeprovisionExecutesSuccessfully() {
	// TODO: Test deprovisioning with no children
	require.True(s.T(), true, "placeholder test")
}

func (s *DeprovisionSignalTestSuite) TestDeprovisionWaitsForChildren() {
	// TODO: Test that deprovision polls until installs/components are deprovisioned
	require.True(s.T(), true, "placeholder test")
}

func (s *DeprovisionSignalTestSuite) TestDeprovisionTimeoutWithChildren() {
	// TODO: Test timeout when children don't deprovision
	require.True(s.T(), true, "placeholder test")
}
