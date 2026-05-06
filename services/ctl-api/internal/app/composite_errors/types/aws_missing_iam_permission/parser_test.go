package aws_missing_iam_permission

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestParser_NoMatchOnUnrelatedInput(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: []byte("Error: creating EC2 Subnet: InvalidSubnet.Range\n"),
	})
	assert.False(t, res.Matched, "non-IAM terraform errors should not match")
}

func TestParser_AccessDeniedExtractsActionPrincipalResource(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: loadFixture(t, "terraform_apply_access_denied.txt"),
	})
	require.True(t, res.Matched)

	e, ok := res.Error.(*Error)
	require.True(t, ok)
	assert.Equal(t, "s3:CreateBucket", e.Action)
	assert.Equal(t, "arn:aws:s3:::acme-prod-assets", e.Resource)
	assert.Equal(t, "arn:aws:iam::123456789012:role/nuon-runner", e.Principal)
	assert.Equal(t, "AccessDenied", e.AWSErrorCode)
	assert.NotEmpty(t, e.RawMessage)
}

func TestParser_UnauthorizedOperationExtractsAction(t *testing.T) {
	p := Parser{}
	res := p.Parse(context.Background(), composite_error.ParseInput{
		Raw: loadFixture(t, "ec2_unauthorized_operation.txt"),
	})
	require.True(t, res.Matched)

	e := res.Error.(*Error)
	assert.Equal(t, "ec2:CreateVpc", e.Action)
	assert.Equal(t, "UnauthorizedOperation", e.AWSErrorCode)
}

func TestErrorOverridesDirectiveToStop(t *testing.T) {
	var ce composite_error.CompositeError = &Error{Action: "s3:CreateBucket"}
	od, ok := ce.(composite_error.ErrorWithDirective)
	require.True(t, ok)
	assert.Equal(t, composite_error.DirectiveStop, od.OverrideDirective().Kind)
}

func TestRender_IncludesPolicyStatement(t *testing.T) {
	e := &Error{
		Action:   "s3:GetObject",
		Resource: "arn:aws:s3:::acme/*",
	}
	r := e.Render(context.Background())

	assert.Contains(t, r.Title, "s3:GetObject")
	require.NotEmpty(t, r.Sections)

	var foundPolicy bool
	for _, s := range r.Sections {
		if s.Heading == "Suggested IAM policy statement" {
			foundPolicy = true
			assert.Contains(t, s.Body, "s3:GetObject")
			assert.Contains(t, s.Body, "arn:aws:s3:::acme/*")
		}
	}
	assert.True(t, foundPolicy)

	// UserActions should include a copy + a docs link + a retry.
	kinds := map[composite_error.UserActionKind]int{}
	for _, ua := range r.UserActions {
		kinds[ua.Kind]++
	}
	assert.Equal(t, 1, kinds[composite_error.UserActionKindCopy])
	assert.Equal(t, 1, kinds[composite_error.UserActionKindLink])
	assert.Equal(t, 1, kinds[composite_error.UserActionKindRetry])
}

func TestRender_PolicyStatementUsesWildcardWhenResourceUnknown(t *testing.T) {
	e := &Error{Action: "ec2:DescribeVpcs"}
	r := e.Render(context.Background())

	var policy string
	for _, s := range r.Sections {
		if s.Heading == "Suggested IAM policy statement" {
			policy = s.Body
		}
	}
	require.NotEmpty(t, policy)
	assert.Contains(t, policy, `"Resource": "*"`)
}

func TestErrorJSONRoundTrip(t *testing.T) {
	original := &Error{
		Action:       "s3:GetObject",
		Resource:     "arn:aws:s3:::a/*",
		Principal:    "arn:aws:iam::1:role/r",
		AWSErrorCode: "AccessDenied",
		RawMessage:   "User is not authorized",
	}
	b, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Error
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, *original, decoded)
}
