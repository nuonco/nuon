package variablesv1

import (
	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("variables", fakeVariables)
	_ = faker.AddProvider("envVars", fakeEnvVars)
	_ = faker.AddProvider("helmValues", fakeHelmValues)
	_ = faker.AddProvider("terraformVariables", fakeTerraformVariables)
	_ = faker.AddProvider("waypointVariables", fakeWaypointVariables)
	_ = faker.AddProvider("intermediateData", fakeIntermediateData)
	_ = faker.AddProvider("installInputs", fakeInstallInputs)
	_ = faker.AddProvider("secrets", fakeSecrets)
}
