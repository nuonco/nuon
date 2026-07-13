package aws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("unable to read fixture %s: %v", name, err)
	}
	return string(b)
}

// parse runs the permission parser directly against raw text, the way the
// registry would once its signal gate passes.
func parse(raw string) compositeerrors.CompositeError {
	return permissionParser{}.Parse(&errparse.ParseContext{Raw: raw})
}

func TestParse_AccessDenied(t *testing.T) {
	ce := parse(readFixture(t, "terraform_apply_access_denied.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}

	e, ok := ce.(*AWSPermissionError)
	if !ok {
		t.Fatalf("expected *AWSPermissionError, got %T", ce)
	}
	if e.Action != "s3:CreateBucket" {
		t.Errorf("action = %q, want s3:CreateBucket", e.Action)
	}
	if e.AWSErrorCode != "AccessDenied" {
		t.Errorf("code = %q, want AccessDenied", e.AWSErrorCode)
	}
	if e.Principal != "arn:aws:iam::123456789012:role/nuon-runner" {
		t.Errorf("principal = %q", e.Principal)
	}
	if e.Resource != "arn:aws:s3:::acme-prod-assets" {
		t.Errorf("resource = %q", e.Resource)
	}
	if ce.Error() != "Missing AWS IAM permission: s3:CreateBucket" {
		t.Errorf("headline = %q", ce.Error())
	}
	if len(ce.Sections()) == 0 {
		t.Error("expected sections, got none")
	}
}

func TestParse_AccessDeniedException_PassRole(t *testing.T) {
	ce := parse(readFixture(t, "iam_passrole_access_denied_exception.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*AWSPermissionError)
	if e.Action != "iam:PassRole" {
		t.Errorf("action = %q, want iam:PassRole", e.Action)
	}
	if e.AWSErrorCode != "AccessDeniedException" {
		t.Errorf("code = %q, want AccessDeniedException", e.AWSErrorCode)
	}
	if e.Principal != "arn:aws:sts::123456789012:assumed-role/nuon-runner/session" {
		t.Errorf("principal = %q", e.Principal)
	}
	if e.Resource != "arn:aws:iam::123456789012:role/acme-task-role" {
		t.Errorf("resource = %q", e.Resource)
	}
}

func TestParse_NoErrorCodePrefix(t *testing.T) {
	// Some SDK clients emit the "is not authorized to perform" sentence with no
	// AccessDenied/Exception code prefix.
	raw := "User: arn:aws:sts::123:assumed-role/foo/bar is not authorized to perform: iam:PassRole on resource: arn:aws:iam::123:role/baz"
	ce := parse(raw)
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*AWSPermissionError)
	if e.Action != "iam:PassRole" {
		t.Errorf("action = %q, want iam:PassRole", e.Action)
	}
	if e.AWSErrorCode != "" {
		t.Errorf("code = %q, want empty (no prefix)", e.AWSErrorCode)
	}
	if e.Resource != "arn:aws:iam::123:role/baz" {
		t.Errorf("resource = %q", e.Resource)
	}
}

func TestParse_UnauthorizedOperation(t *testing.T) {
	ce := parse(readFixture(t, "ec2_unauthorized_operation.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*AWSPermissionError)
	if e.Action != "ec2:CreateVpc" {
		t.Errorf("action = %q, want ec2:CreateVpc", e.Action)
	}
	if e.AWSErrorCode != "UnauthorizedOperation" {
		t.Errorf("code = %q, want UnauthorizedOperation", e.AWSErrorCode)
	}
}

// TestParse_S3AccessDenied_PermissionsBoundary uses the real error_output the
// runner captured from a broken-provision deploy: plain log-capture lines (no
// terraform box-drawing), a quoted resource ARN, and the "explicit deny in a
// permissions boundary" phrasing. It guards the quote-stripping in
// cleanResource, since a quoted ARN would otherwise produce a malformed IAM
// policy statement in the "How to fix" section.
func TestParse_S3AccessDenied_PermissionsBoundary(t *testing.T) {
	ce := parse(readFixture(t, "s3_access_denied_permissions_boundary.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*AWSPermissionError)
	if e.Action != "s3:CreateBucket" {
		t.Errorf("action = %q, want s3:CreateBucket", e.Action)
	}
	if e.AWSErrorCode != "AccessDenied" {
		t.Errorf("code = %q, want AccessDenied", e.AWSErrorCode)
	}
	if e.Resource != "arn:aws:s3:::inlgckaypxrqqlbs1t7axp26p8-nuon-clickhouse" {
		t.Errorf("resource = %q, want unquoted ARN", e.Resource)
	}
	if strings.ContainsAny(e.Resource, `"'`) {
		t.Errorf("resource %q still carries quotes", e.Resource)
	}
	// The generated policy statement must embed the clean ARN, not a quoted one.
	policy := ""
	for _, s := range e.Sections() {
		if s.Heading == "IAM policy statement" {
			policy = s.Body
		}
	}
	if !strings.Contains(policy, `"arn:aws:s3:::inlgckaypxrqqlbs1t7axp26p8-nuon-clickhouse"`) {
		t.Errorf("IAM policy statement section missing clean resource ARN: %q", policy)
	}
	if strings.Contains(policy, `\"arn:aws:s3`) {
		t.Errorf("IAM policy statement section has a doubly-quoted resource: %q", policy)
	}
}

func TestParse_NoMatch(t *testing.T) {
	cases := []string{
		"",
		"Error: creating EC2 VPC: InvalidParameterValue: bad CIDR",
		"plan job failed",
	}
	for _, in := range cases {
		if ce := parse(in); ce != nil {
			t.Errorf("parse(%q) = %v, want nil", in, ce)
		}
	}
}

func TestParse_HintsSkipAutoRetry(t *testing.T) {
	ce := parse(readFixture(t, "terraform_apply_access_denied.txt"))
	hp, ok := ce.(compositeerrors.HintsProvider)
	if !ok {
		t.Fatal("expected AWSPermissionError to provide hints")
	}
	if !hp.Hints().SkipAutoRetry() {
		t.Error("expected skip_auto_retry hint on a missing-permission error")
	}
}

func TestSections_FullyPopulated(t *testing.T) {
	e := &AWSPermissionError{
		Action:       "s3:CreateBucket",
		Resource:     "arn:aws:s3:::acme-prod-assets",
		Principal:    "arn:aws:iam::123456789012:role/nuon-runner",
		AWSErrorCode: "AccessDenied",
		RawMessage:   "AccessDenied: ... is not authorized to perform: s3:CreateBucket",
	}

	headings := map[string]string{}
	for _, s := range e.Sections() {
		headings[s.Heading] = s.Body
	}

	for _, want := range []string{"Why", "AWS response", "Context", "How to fix", "IAM policy statement"} {
		if _, ok := headings[want]; !ok {
			t.Errorf("missing section %q; got sections %v", want, headings)
		}
	}

	policy := headings["IAM policy statement"]
	if !strings.Contains(policy, "s3:CreateBucket") || !strings.Contains(policy, "arn:aws:s3:::acme-prod-assets") {
		t.Errorf("IAM policy statement section missing action/resource: %q", policy)
	}
	if !strings.Contains(headings["Context"], "arn:aws:iam::123456789012:role/nuon-runner") {
		t.Errorf("Context section missing principal: %q", headings["Context"])
	}
}

func TestSections_MinimalOmitsOptionalSections(t *testing.T) {
	e := &AWSPermissionError{}

	headings := map[string]bool{}
	for _, s := range e.Sections() {
		headings[s.Heading] = true
	}
	if headings["AWS response"] || headings["Context"] || headings["How to fix"] {
		t.Errorf("expected only the Why section for an empty error, got %v", headings)
	}
}

// TestRegistry_GatesAndDispatches exercises the registry path end to end: the
// signal gate must let a matching AWS error through and reject unrelated text,
// and an unknown tool/provider must fail open so the parser still runs.
func TestRegistry_GatesAndDispatches(t *testing.T) {
	raw := readFixture(t, "terraform_apply_access_denied.txt")

	ce := errparse.Parse(&errparse.ParseContext{Raw: raw})
	if ce == nil {
		t.Fatal("expected registry to dispatch to the AWS permission parser")
	}
	if _, ok := ce.(*AWSPermissionError); !ok {
		t.Fatalf("expected *AWSPermissionError, got %T", ce)
	}

	if got := errparse.Parse(&errparse.ParseContext{Raw: "plan job failed"}); got != nil {
		t.Errorf("expected nil for unrelated text, got %v", got)
	}
}
