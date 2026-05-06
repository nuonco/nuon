package terraform_apply_failed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return b
}

func TestParser_NoMatchOnNonTerraformInput(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: []byte("some random non-terraform error"),
	})
	assert.False(t, res.Matched)
}

func TestParser_ExtractsSingleDiagnostic(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw:        loadFixture(t, "terraform_apply_subnet_failure.txt"),
		Invocation: composite_error.InvocationContext{OwnerType: "install_apply_step"},
	})
	require.True(t, res.Matched)

	e, ok := res.Error.(*Error)
	require.True(t, ok)
	assert.Equal(t, "apply", e.Stage)
	require.Len(t, e.Diagnostics, 1)

	d := e.Diagnostics[0]
	assert.Contains(t, d.Summary, "creating EC2 Subnet")
	assert.Contains(t, d.Summary, "InvalidSubnet.Range")
	assert.Equal(t, "module.vpc.aws_subnet.public[2]", d.Resource)
	assert.Equal(t, ".terraform/modules/vpc/main.tf", d.SourceFile)
	assert.Equal(t, 218, d.SourceLine)
}

func TestParser_ExtractsMultipleDiagnostics(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: loadFixture(t, "terraform_apply_multiple_diagnostics.txt"),
	})
	require.True(t, res.Matched)

	e, ok := res.Error.(*Error)
	require.True(t, ok)
	require.Len(t, e.Diagnostics, 2)

	assert.Contains(t, e.Diagnostics[0].Summary, "EntityAlreadyExists")
	assert.Equal(t, "module.eks.aws_iam_role.cluster", e.Diagnostics[0].Resource)
	assert.Equal(t, ".terraform/modules/eks/iam.tf", e.Diagnostics[0].SourceFile)
	assert.Equal(t, 12, e.Diagnostics[0].SourceLine)

	assert.Contains(t, e.Diagnostics[1].Summary, "VPCIdNotSpecified")
	assert.Equal(t, "module.eks.aws_security_group.cluster", e.Diagnostics[1].Resource)
}

func TestParser_StageFallsBackFromExtra(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: loadFixture(t, "terraform_apply_subnet_failure.txt"),
		Invocation: composite_error.InvocationContext{
			Extra: map[string]any{"terraform_stage": "destroy"},
		},
	})
	require.True(t, res.Matched)
	e := res.Error.(*Error)
	assert.Equal(t, "destroy", e.Stage)
}

func TestRender_SingleDiagnosticIncludesResource(t *testing.T) {
	e := &Error{
		Stage: "apply",
		Diagnostics: []Diagnostic{{
			Summary:    "creating EC2 Subnet: InvalidSubnet.Range",
			Resource:   "module.vpc.aws_subnet.public[2]",
			SourceFile: ".terraform/modules/vpc/main.tf",
			SourceLine: 218,
		}},
	}
	r := e.Render(context.Background())
	assert.Contains(t, r.Title, "Terraform apply failed")
	require.Len(t, r.Sections, 1)
	assert.Contains(t, r.Sections[0].Body, "module.vpc.aws_subnet.public[2]")
	assert.Contains(t, r.Sections[0].Body, "line 218")
}

func TestRender_MultipleDiagnosticsHaveSummaryList(t *testing.T) {
	e := &Error{
		Stage: "apply",
		Diagnostics: []Diagnostic{
			{Summary: "first"},
			{Summary: "second"},
		},
	}
	r := e.Render(context.Background())
	assert.Contains(t, r.Title, "2 errors")
	assert.True(t, strings.HasPrefix(r.Summary, "• "))
}

func TestErrorJSONRoundTrip(t *testing.T) {
	original := &Error{
		Stage: "apply",
		Diagnostics: []Diagnostic{{
			Summary:    "boom",
			Resource:   "aws_vpc.main",
			SourceFile: "main.tf",
			SourceLine: 7,
			Raw:        "│ Error: boom",
		}},
	}
	b, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Error
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, *original, decoded)
}
