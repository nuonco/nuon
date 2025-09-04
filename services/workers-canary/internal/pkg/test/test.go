package test

import "github.com/powertoolsdev/mono/services/workers-canary/internal/app"

// CanaryTest interface defines the contract for all canary tests
type CanaryTests interface {
	// Setup prepares the environment for the CanaryTests
	Setup() error

	// RunTests executes the one or more tests and returns their results
	ExecTests() ([]*app.TestRun, error)

	// Teardown cleans up any resources used during the test
	Teardown() error

	// GetProperties returns common properties of the test
	GetProperties() CanaryTestProperties
}

// CanaryTestProperties defines common properties every test should have
type CanaryTestProperties struct {
	CanaryID         string `json:"canary_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	AppConfigPath    string `json:"app_config_path"`
	AppConfigGithash string `json:"app_config_githash"`
	MonoGithash      string `json:"mono_githash"`
}
