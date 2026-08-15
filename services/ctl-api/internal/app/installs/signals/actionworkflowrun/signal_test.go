package actionworkflowrun

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type SignalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestSignalTestSuite(t *testing.T) {
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) TestExecute() {
	// TODO: implement test
	s.T().Skip("not yet implemented")
}

func (s *SignalTestSuite) TestSupportedImageActionPlatform() {
	for _, tc := range []struct {
		name     string
		platform app.AppRunnerType
		want     bool
	}{
		{name: "aws is supported", platform: app.AppRunnerTypeAWS, want: true},
		{name: "local is supported for development", platform: app.AppRunnerTypeLocal, want: true},
		{name: "azure is rejected", platform: app.AppRunnerTypeAzure, want: false},
		{name: "gcp is rejected", platform: app.AppRunnerTypeGCP, want: false},
		{name: "aws ecs is rejected", platform: app.AppRunnerTypeAWSECS, want: false},
		{name: "aws eks is rejected", platform: app.AppRunnerTypeAWSEKS, want: false},
		{name: "azure aks is rejected", platform: app.AppRunnerTypeAzureAKS, want: false},
		{name: "gcp gke is rejected", platform: app.AppRunnerTypeGCPGKE, want: false},
		{name: "empty is rejected", platform: app.AppRunnerType(""), want: false},
	} {
		s.Run(tc.name, func() {
			s.Equal(tc.want, supportedImageActionPlatform(tc.platform))
		})
	}
}
