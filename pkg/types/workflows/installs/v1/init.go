package installsv1

import (
	"github.com/go-faker/faker/v4"
)

//nolint:gochecknoinits
func init() {
	_ = faker.AddProvider("installTerraformOutputs", fakeInstallTerraformOutputs)
}
