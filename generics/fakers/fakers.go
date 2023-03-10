package fakers

import (
	"github.com/go-faker/faker/v4"
)

func Register() {
	_ = faker.AddProvider("shortID", fakeShortID)
	_ = faker.AddProvider("buildConfig", fakeBuildConfig)
	_ = faker.AddProvider("deployConfig", fakeDeployConfig)
}
