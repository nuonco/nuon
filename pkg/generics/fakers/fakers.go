package fakers

import (
	"github.com/go-faker/faker/v4"
)

func Register() {
	_ = faker.AddProvider("shortID", fakeShortID)

	// components
	_ = faker.AddProvider("buildConfig", fakeBuildConfig)
	_ = faker.AddProvider("deployConfig", fakeDeployConfig)

	// plans
	_ = faker.AddProvider("planConfigs", fakePlanConfigs)
	_ = faker.AddProvider("sandboxInputAccountSettings", fakeSandboxInputAccountSettings)
	_ = faker.AddProvider("envVars", fakeEnvVars)
	_ = faker.AddProvider("waypointVariables", fakeWaypointVariables)

	// pkg/pipeline
	_ = faker.AddProvider("pipelineCallbackFn", fakePipelineCallbackFn)
	_ = faker.AddProvider("pipelineExecFn", fakePipelineExecFn)

	// api fakers
	_ = faker.AddProvider("apiInstallAWSSettings", fakeAPIInstallAWSSettings)
}
