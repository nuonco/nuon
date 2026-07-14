package errparse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/all"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/errparsetest"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// This external test registers parsers solely through the errparse/all
// manifest, the same wiring production uses, so the two can never drift. It runs
// fixtures through the contract runner, which dispatches against the real
// registry and enforces both the layer contract and the persistence
// round-trip. Cases match on the winning composite error's Type() rather than a
// concrete parser type so no parser package is imported directly here.
const (
	awsPermissionType  compositeerrors.Type = "terraform.aws_permission"
	genericType        compositeerrors.Type = "generic"
	terraformType      compositeerrors.Type = "terraform.error"
	terraformStateLock compositeerrors.Type = "terraform.state_lock"
	helmNameInUseType  compositeerrors.Type = "helm.name_in_use"
)

func readFixture(t *testing.T, tool, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(tool, "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s/%s: %v", tool, name, err)
	}
	return string(b)
}

func TestDefaultRegistry_Contract(t *testing.T) {
	awsTerraformBlob := "Error: creating S3 Bucket (acme): AccessDenied: User: " +
		"arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme"
	awsHelmBlob := "unable to upgrade helm release: creating S3 Bucket (acme): AccessDenied: " +
		"User: arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme"
	helmNameInUse := "job step errored unable to execute job: unable to upgrade helm release: " +
		"cannot reuse a name that is still in use"

	errparsetest.Run(t, []errparsetest.Case{
		{
			Name:     "provider layer beats generic",
			Raw:      readFixture(t, "aws", "terraform_apply_access_denied.txt"),
			WantType: awsPermissionType,
		},
		{
			Name:     "generic catches unclassified",
			Raw:      "helm upgrade failed: context deadline exceeded",
			WantType: genericType,
		},
		{
			Name:     "terraform tool layer beats generic",
			Raw:      readFixture(t, "terraform", "invalid_reference.txt"),
			Tool:     errparse.ToolTerraform,
			WantType: terraformType,
		},
		{
			Name:     "provider layer beats terraform tool layer",
			Raw:      awsTerraformBlob,
			Tool:     errparse.ToolTerraform,
			WantType: awsPermissionType,
		},
		{
			Name:     "state-lock parser beats terraform catch-all",
			Raw:      readFixture(t, "terraform", "state_lock.txt"),
			Tool:     errparse.ToolTerraform,
			WantType: terraformStateLock,
		},
		{
			Name:     "helm tool layer beats generic",
			Raw:      helmNameInUse,
			Tool:     errparse.ToolHelm,
			WantType: helmNameInUseType,
		},
		{
			Name:     "provider layer beats helm tool layer",
			Raw:      awsHelmBlob,
			Tool:     errparse.ToolHelm,
			WantType: awsPermissionType,
		},
	})
}
