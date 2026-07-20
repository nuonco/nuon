package awaitinstallstackversionrun

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

func TestShouldCreateManagedAWSCloudFormationStack(t *testing.T) {
	connectionID := "awc_test"
	tests := map[string]struct {
		create  bool
		install *app.Install
		appCfg  *app.AppConfig
		want    bool
	}{
		"verified connection selection": {
			create:  true,
			install: &app.Install{AWSAccount: &app.AWSAccount{AWSAccountConnectionID: &connectionID}},
			appCfg:  &app.AppConfig{RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS}},
			want:    true,
		},
		"manual AWS install": {
			create:  true,
			install: &app.Install{AWSAccount: &app.AWSAccount{}},
			appCfg:  &app.AppConfig{RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS}},
		},
		"sandbox install": {
			create: true,
			install: &app.Install{
				SandboxMode: sql.NullBool{Bool: true, Valid: true},
				AWSAccount:  &app.AWSAccount{AWSAccountConnectionID: &connectionID},
			},
			appCfg: &app.AppConfig{RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS}},
		},
		"non-AWS install": {
			create:  true,
			install: &app.Install{AWSAccount: &app.AWSAccount{AWSAccountConnectionID: &connectionID}},
			appCfg:  &app.AppConfig{RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAzure}},
		},
		"reprovision": {
			install: &app.Install{AWSAccount: &app.AWSAccount{AWSAccountConnectionID: &connectionID}},
			appCfg:  &app.AppConfig{RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := shouldCreateManagedAWSCloudFormationStack(tt.create, tt.install, tt.appCfg); got != tt.want {
				t.Fatalf("shouldCreateManagedAWSCloudFormationStack() = %v, want %v", got, tt.want)
			}
		})
	}
}
