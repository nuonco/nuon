//go:build integrationlocal

package runner

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Run_Int(t *testing.T) {
	t.Parallel()

	// plan := &planv1.TerraformPlan{
	// 	Id:      "testid",
	// 	RunType: planv1.RunType_RUN_TYPE_APPLY,
	// 	Module: &planv1.Object{
	// 		Bucket: "nuon-sandboxes",
	// 		Key:    "sandboxes/empty_0.8.8.tar.gz",
	// 		Region: "us-west-2",
	// 	},
	// 	Backend: &planv1.Object{
	// 		Bucket: "jdt-test",
	// 		Key:    "tf-runner/state.json",
	// 		Region: "us-west-2",
	// 	},
	// 	Vars: map[string]*anypb.Any{},
	// }
	//
	// bs, err := proto.Marshal(plan)
	// assert.NoError(t, err)
	// assert.NoError(t, os.WriteFile("request.proto", bs, 0o600))
	// t.FailNow()

	r, err := New(
		validator.New(),
		WithPlan(&planv1.PlanRef{
			Bucket:              "jdt-test",
			BucketKey:           "request.proto",
			BucketAssumeRoleArn: "arn:aws:iam::649224399387:role/aws-reserved/sso.amazonaws.com/us-east-2/AWSReservedSSO_NuonAdmin_c083f7fead01d64e",
		}),
	)
	assert.NoError(t, err)

	m, err := r.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"test_number": "1", "test_string": "test_string"}, m)
}
