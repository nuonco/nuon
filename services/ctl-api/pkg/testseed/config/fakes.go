package config

import (
	"github.com/go-faker/faker/v4"
)

// init registers all faker providers for config structures.
// These providers are used by generics.GetFakeObj to generate valid test configs.
func init() {
	// Main config providers
	_ = faker.AddProvider("appConfig", fakeAppConfig)
	_ = faker.AddProvider("sandboxConfig", fakeSandboxConfig)
	_ = faker.AddProvider("runnerConfig", fakeRunnerConfig)

	// Additional config providers
	_ = faker.AddProvider("inputConfig", fakeInputConfig)
	_ = faker.AddProvider("permissionsConfig", fakePermissionsConfig)
	_ = faker.AddProvider("policiesConfig", fakePoliciesConfig)
	_ = faker.AddProvider("secretsConfig", fakeSecretsConfig)
	_ = faker.AddProvider("breakGlassConfig", fakeBreakGlassConfig)
	_ = faker.AddProvider("stackConfig", fakeStackConfig)

	// Component providers
	_ = faker.AddProvider("terraformComponent", fakeTerraformComponent)
	_ = faker.AddProvider("helmComponent", fakeHelmComponent)
	_ = faker.AddProvider("dockerComponent", fakeDockerComponent)
	_ = faker.AddProvider("k8sManifestComponent", fakeKubernetesManifestComponent)
	_ = faker.AddProvider("jobComponent", fakeJobComponent)
	_ = faker.AddProvider("externalImageComponent", fakeExternalImageComponent)
}
