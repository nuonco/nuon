package fakers

import (
	"reflect"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func fakeSandboxInputAccountSettings(v reflect.Value) (interface{}, error) {
	return &planv1.SandboxInput_Aws{
		Aws: &planv1.AWSSettings{
			Region:    "us-west-2",
			AccountId: "543425801867",
			RoleArn:   "arn:aws:iam::543425801867:role/nuon-workers-canary-prod-install-access",
		},
	}, nil
}
