package domains

import (
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/powertoolsdev/mono/pkg/shortid"
)

func fakeAppID(v reflect.Value) (interface{}, error) {
	return shortid.NewNanoID("app"), nil
}

func fakeArtifactID(v reflect.Value) (interface{}, error) {
	return NewArtifactID(), nil
}

func fakeAWSSettingsID(v reflect.Value) (interface{}, error) {
	return NewAWSAccountID(), nil
}

func fakeBuildID(v reflect.Value) (interface{}, error) {
	return NewBuildID(), nil
}

func fakeCanaryID(v reflect.Value) (interface{}, error) {
	return NewCanaryID(), nil
}

func fakeComponentID(v reflect.Value) (interface{}, error) {
	return NewComponentID(), nil
}

func fakeDeploymentID(v reflect.Value) (interface{}, error) {
	return NewDeploymentID(), nil
}

func fakeDeployID(v reflect.Value) (interface{}, error) {
	return NewDeployID(), nil
}

func fakeInstallID(v reflect.Value) (interface{}, error) {
	return NewInstallID(), nil
}

func fakeInstanceID(v reflect.Value) (interface{}, error) {
	return NewInstanceID(), nil
}

func fakeOrgID(v reflect.Value) (interface{}, error) {
	return NewOrgID(), nil
}

func fakeSandboxID(v reflect.Value) (interface{}, error) {
	return NewSandboxID(), nil
}

func fakeSecretID(v reflect.Value) (interface{}, error) {
	return shortid.NewNanoID("sec"), nil
}

func fakeUserID(v reflect.Value) (interface{}, error) {
	return NewUserID(), nil
}

func init() {
	_ = faker.AddProvider("appID", fakeAppID)
	_ = faker.AddProvider("artifactID", fakeArtifactID)
	_ = faker.AddProvider("awsSettingsID", fakeAWSSettingsID)
	_ = faker.AddProvider("buildID", fakeBuildID)
	_ = faker.AddProvider("canaryID", fakeCanaryID)
	_ = faker.AddProvider("componentID", fakeComponentID)
	_ = faker.AddProvider("deployID", fakeDeployID)
	_ = faker.AddProvider("deploymentID", fakeDeploymentID)
	_ = faker.AddProvider("installID", fakeInstallID)
	_ = faker.AddProvider("instanceID", fakeInstanceID)
	_ = faker.AddProvider("orgID", fakeOrgID)
	_ = faker.AddProvider("orgID", fakeOrgID)
	_ = faker.AddProvider("sandboxID", fakeSandboxID)
	_ = faker.AddProvider("secretID", fakeSecretID)
	_ = faker.AddProvider("userID", fakeUserID)
}
