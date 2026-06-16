package deployerrors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("unable to read fixture %s: %v", name, err)
	}
	return string(b)
}

func TestParse_AccessDenied(t *testing.T) {
	ce := Parse(readFixture(t, "terraform_apply_access_denied.txt"))
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
	ce := Parse(readFixture(t, "iam_passrole_access_denied_exception.txt"))
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
	ce := Parse(raw)
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
	ce := Parse(readFixture(t, "ec2_unauthorized_operation.txt"))
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

	for _, want := range []string{"Why", "AWS response", "Context", "How to fix"} {
		if _, ok := headings[want]; !ok {
			t.Errorf("missing section %q; got sections %v", want, headings)
		}
	}

	// The "How to fix" section must embed a valid IAM policy granting the action.
	fix := headings["How to fix"]
	if !strings.Contains(fix, "s3:CreateBucket") || !strings.Contains(fix, "arn:aws:s3:::acme-prod-assets") {
		t.Errorf("How to fix section missing action/resource: %q", fix)
	}
	if !strings.Contains(headings["Context"], "arn:aws:iam::123456789012:role/nuon-runner") {
		t.Errorf("Context section missing principal: %q", headings["Context"])
	}
}

func TestSections_MinimalOmitsOptionalSections(t *testing.T) {
	e := &AWSPermissionError{} // no action/resource/principal/raw

	headings := map[string]bool{}
	for _, s := range e.Sections() {
		headings[s.Heading] = true
	}
	if headings["AWS response"] || headings["Context"] || headings["How to fix"] {
		t.Errorf("expected only the Why section for an empty error, got %v", headings)
	}
}

func TestParse_NoMatch(t *testing.T) {
	cases := []string{
		"",
		"Error: creating EC2 VPC: InvalidParameterValue: bad CIDR",
		"plan job failed",
	}
	for _, in := range cases {
		if ce := Parse(in); ce != nil {
			t.Errorf("Parse(%q) = %v, want nil", in, ce)
		}
	}
}
