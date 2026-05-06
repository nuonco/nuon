package composite_errors_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"

	// Side-effect imports register every built-in CompositeError type.
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/register"

	awsperm "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/aws_missing_iam_permission"
	tferr "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/terraform_apply_failed"
	unknown "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/unknown_error"
)

// loadFixture pulls a testdata file from one of the type subpackages.
func loadFixture(t *testing.T, typeDir, name string) []byte {
	t.Helper()
	p := filepath.Join("types", typeDir, "testdata", name)
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return b
}

// pipeline returns a Pipeline wired against the real catalog and the
// unknown_error fallback builder.
func pipeline() *composite_error.Pipeline {
	return composite_error.NewPipeline(catalog.ParsersForContext, unknown.Build)
}

// TestEndToEnd_AWSPermissionWinsOverGenericTerraform verifies that when both
// the broad terraform parser and the cross-cutting AWS parser match the same
// input, the more specific (parser-declared) AWS one becomes the primary and
// the broader terraform diagnostic becomes a secondary.
//
// In v1, "specificity" is determined by parser order at the same registration
// level. Both parsers register at the "terraform" subtree, so order matters —
// AWS is intentionally registered first by the type package's init.
func TestEndToEnd_AWSPermissionWinsOverGenericTerraform(t *testing.T) {
	raw := loadFixture(t, "aws_missing_iam_permission", "terraform_apply_access_denied.txt")

	res := pipeline().Parse(context.Background(), "terraform/apply", composite_error.ParseInput{
		Raw: raw,
		Invocation: composite_error.InvocationContext{
			OwnerType:     "install_apply_step",
			CloudPlatform: "aws",
		},
	})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, awsperm.Type, res.Primary.Error.Type(), "AWS perm parser should claim the input")

	// The terraform broad parser also matches and becomes a secondary.
	require.Len(t, res.Secondaries, 1)
	assert.Equal(t, tferr.Type, res.Secondaries[0].Error.Type())
}

// TestEndToEnd_GenericTerraformDiagnosticOnNonAWSError verifies that a
// terraform error which is not an AWS permission failure falls through to
// the broad terraform parser as primary, with no secondaries.
func TestEndToEnd_GenericTerraformDiagnosticOnNonAWSError(t *testing.T) {
	raw := loadFixture(t, "terraform_apply_failed", "terraform_apply_subnet_failure.txt")

	res := pipeline().Parse(context.Background(), "terraform/apply", composite_error.ParseInput{
		Raw: raw,
	})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, tferr.Type, res.Primary.Error.Type())
	assert.Empty(t, res.Secondaries)

	te := res.Primary.Error.(*tferr.Error)
	require.Len(t, te.Diagnostics, 1)
	assert.Contains(t, te.Diagnostics[0].Summary, "InvalidSubnet.Range")
}

// TestEndToEnd_UnknownFallback verifies the safety-net behaviour: input that
// no parser claims yields a single unknown_error primary.
func TestEndToEnd_UnknownFallback(t *testing.T) {
	res := pipeline().Parse(context.Background(), "kubernetes/rollout", composite_error.ParseInput{
		Raw: []byte("a wild error appeared"),
	})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, unknown.Type, res.Primary.Error.Type())
	assert.Empty(t, res.Secondaries)
}

// TestEndToEnd_HydrateRoundTrip verifies the persistence round-trip:
// marshal a typed error → store as JSON → hydrate via the catalog → render.
//
// This is the contract the DB layer will rely on once helpers.Record() lands.
func TestEndToEnd_HydrateRoundTrip(t *testing.T) {
	original := &awsperm.Error{
		Action:    "ec2:CreateVpc",
		Resource:  "*",
		Principal: "arn:aws:iam::1:role/runner",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	got, err := catalog.Hydrate(awsperm.Type, data)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, awsperm.Type, got.Type())

	rendered := got.Render(context.Background())
	assert.Contains(t, rendered.Title, "ec2:CreateVpc")
}

// TestEndToEnd_PipelineHonorsParseContextScoping verifies the AWS perm parser
// does not run for kubernetes/rollout (it only opts into terraform, helm, and
// runner/job).
func TestEndToEnd_PipelineHonorsParseContextScoping(t *testing.T) {
	// An input that WOULD match the AWS parser if it ran, but kubernetes/rollout
	// is outside its declared subtrees, so it should fall back to unknown.
	res := pipeline().Parse(context.Background(), "kubernetes/rollout", composite_error.ParseInput{
		Raw: []byte("AccessDenied: User: arn:aws:iam::1:role/r is not authorized to perform: s3:GetObject on resource: arn:aws:s3:::x"),
	})

	assert.Equal(t, unknown.Type, res.Primary.Error.Type())
}
