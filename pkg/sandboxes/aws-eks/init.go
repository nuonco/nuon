package awseks

import (
	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("sandboxAwsEksFakeOutputs", fakeOutputs)
	_ = faker.AddProvider("stringSliceAsInt", fakeStringSliceAsInt)
	_ = faker.AddProvider("domain", fakeDomain)
}
