//go:build integrationlocal

// TODO(jdt): get to where this is just an integration test

package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

// single test to setup workspace end-to-end
func TestWorkspace_Setup_Int(t *testing.T) {
	t.Parallel()

	w, err := NewWorkspace(
		validator.New(),
		WithID(t.Name()),
		// doesn't matter for this test
		WithBackendBucket(&planv1.Object{
			Bucket: "testbucket",
			Key:    "testkey",
			Region: "us-east-1",
		}),
		WithModuleBucket(&planv1.Object{
			Bucket: "nuon-sandboxes",
			Key:    "sandboxes/empty_0.8.33.tar.gz",
			Region: "us-west-2",
			AssumeRoleDetails: &planv1.AssumeRoleDetails{
				AssumeArn: "arn:aws:iam::649224399387:role/jdt-terraform-exec-test",
			},
		}),
		WithVars(map[string]interface{}{"test": "vars"}),
	)
	assert.NoError(t, err)

	err = w.Setup(context.Background())
	assert.NoError(t, err)

	assert.DirExists(t, w.workingDir)
	for _, f := range []string{"backend.json", "nuon.tfvars.json", "output.tf", "provider.tf", "versions.tf"} {
		assert.FileExists(t, filepath.Join(w.workingDir, f))
	}
	assert.NoError(t, w.Cleanup())
	assert.NoDirExists(t, filepath.Join(os.TempDir(), w.ID))
}
