//go:build integrationlocal

package runner

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/plan/v1"
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
	// 		AssumeRoleDetails: &planv1.AssumeRoleDetails{
	// 			AssumeArn: "arn:aws:iam::649224399387:role/jdt-terraform-exec-test",
	// 		},
	// 	},
	// 	Backend: &planv1.Object{
	// 		Bucket: "jdt-test",
	// 		Key:    "tf-runner/state.json",
	// 		Region: "us-west-2",
	// 		AssumeRoleDetails: &planv1.AssumeRoleDetails{
	// 			AssumeArn: "arn:aws:iam::649224399387:role/jdt-terraform-exec-test",
	// 		},
	// 	},
	// 	Vars: &structpb.Struct{},
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
			BucketAssumeRoleArn: "arn:aws:iam::649224399387:role/jdt-terraform-exec-test",
		}),
	)
	assert.NoError(t, err)

	m, err := r.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, map[string]interface{}{
		"test_number": float64(1),
		"test_string": "test_string",
		"test_map":    map[string]interface{}{"number": float64(1), "string": "a"},
	}, m)
}
