package config

import (
	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("appConfig", fakeAppConfig)
	_ = faker.AddProvider("sandboxConfig", fakeSandboxConfig)
	_ = faker.AddProvider("runnerConfig", fakeRunnerConfig)

	_ = faker.AddProvider("inputConfig", fakeInputConfig)
	_ = faker.AddProvider("permissionsConfig", fakePermissionsConfig)
	_ = faker.AddProvider("policiesConfig", fakePoliciesConfig)
	_ = faker.AddProvider("secretsConfig", fakeSecretsConfig)
	_ = faker.AddProvider("breakGlassConfig", fakeBreakGlassConfig)
	_ = faker.AddProvider("stackConfig", fakeStackConfig)

	_ = faker.AddProvider("terraformComponent", fakeTerraformComponent)
	_ = faker.AddProvider("helmComponent", fakeHelmComponent)
	_ = faker.AddProvider("dockerComponent", fakeDockerComponent)
	_ = faker.AddProvider("k8sManifestComponent", fakeKubernetesManifestComponent)
	_ = faker.AddProvider("jobComponent", fakeJobComponent)
	_ = faker.AddProvider("externalImageComponent", fakeExternalImageComponent)
}
