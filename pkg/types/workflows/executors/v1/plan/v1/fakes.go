package planv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
)

func init() {
	_ = faker.AddProvider("sandboxInputAccountSettings", fakeSandboxInputAccountSettings)
}

func fakeSandboxInputAccountSettings(v reflect.Value) (interface{}, error) {
	return &SandboxInput_Aws{
		Aws: &AWSSettings{
			Region:    "us-west-2",
			AccountId: "543425801867",
			RoleArn:   "arn:aws:iam::543425801867:role/nuon-workers-canary-prod-install-access",
		},
	}, nil
}
