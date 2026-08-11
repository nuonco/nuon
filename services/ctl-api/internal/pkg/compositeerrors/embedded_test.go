package compositeerrors

import (
	"strings"
	"testing"
)

// unmarshalableError carries a field json.Marshal cannot encode, so New() must
// surface the marshal failure instead of persisting a record with a null typed
// payload.
type unmarshalableError struct {
	Ch chan int `json:"ch"`
}

func (unmarshalableError) Error() string       { return "bad" }
func (unmarshalableError) Type() Type          { return "test.unmarshalable" }
func (unmarshalableError) Severity() Severity  { return SeverityError }
func (unmarshalableError) Sections() []Section { return nil }

func TestNew_ReturnsErrorOnUnmarshalablePayload(t *testing.T) {
	d, err := New(unmarshalableError{})
	if err == nil {
		t.Fatal("expected an error for an unmarshalable payload")
	}
	if d != nil {
		t.Fatalf("expected nil record on error, got %+v", d)
	}
}

// sectionedError returns a caller-owned slice from Sections() so the test can
// confirm New() detaches its copy from the source.
type sectionedError struct{ sections []Section }

func (sectionedError) Error() string         { return "boom" }
func (sectionedError) Type() Type            { return "test.sectioned" }
func (sectionedError) Severity() Severity    { return SeverityError }
func (e sectionedError) Sections() []Section { return e.sections }

func TestNew_FreezesHintsAndSectionsFromSource(t *testing.T) {
	t.Run("hints detached from a shared source map", func(t *testing.T) {
		shared := NewHints().WithSkipAutoRetry()
		d, err := New(hintedError{hints: shared})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !d.Hints.SkipAutoRetry() {
			t.Fatal("expected the hint to be captured")
		}
		shared[HintSkipAutoRetry] = "false"
		if !d.Hints.SkipAutoRetry() {
			t.Fatal("persisted hints must not observe a mutation of the source map")
		}
	})

	t.Run("sections detached from the source slice", func(t *testing.T) {
		src := []Section{CodeSection("Output", "original")}
		d, err := New(sectionedError{sections: src})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		src[0].Body = "mutated"
		if d.Sections[0].Body != "original" {
			t.Fatalf("persisted section must not observe a mutation of the source slice, got %q", d.Sections[0].Body)
		}
	})
}

type sensitiveError struct {
	Output string         `json:"output"`
	Nested map[string]any `json:"nested,omitempty"`
	hints  Hints
}

func (e sensitiveError) Error() string       { return e.Output }
func (sensitiveError) Type() Type            { return "test.sensitive" }
func (sensitiveError) Severity() Severity    { return SeverityError }
func (e sensitiveError) Sections() []Section { return []Section{CodeSection("Output", e.Output)} }
func (e sensitiveError) Hints() Hints        { return e.hints }

func TestRedactDiagnosticSecrets(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"nuon token query": {
			input: "https://api.example.com/state?workspace_id=tfw123&token=tok_secret&operation=plan",
			want:  "https://api.example.com/state?workspace_id=tfw123&token=[REDACTED]&operation=plan",
		},
		"case insensitive multiple values": {
			input: "https://example.com?API_KEY=key-secret&Client-Secret=client-secret",
			want:  "https://example.com?API_KEY=[REDACTED]&Client-Secret=[REDACTED]",
		},
		"aws signed URL": {
			input: "https://s3.example.com/object?X-Amz-Credential=credential&X-Amz-Security-Token=session-token&X-Amz-Signature=signature",
			want:  "https://s3.example.com/object?X-Amz-Credential=[REDACTED]&X-Amz-Security-Token=[REDACTED]&X-Amz-Signature=[REDACTED]",
		},
		"azure SAS URL": {
			input: "https://account.blob.core.windows.net/container?sv=2025-01-05&sig=azure-signature&sp=r",
			want:  "https://account.blob.core.windows.net/container?sv=2025-01-05&sig=[REDACTED]&sp=r",
		},
		"gcp signed URL": {
			input: "https://storage.googleapis.com/object?X-Goog-Credential=credential&X-Goog-Signature=gcp-signature",
			want:  "https://storage.googleapis.com/object?X-Goog-Credential=[REDACTED]&X-Goog-Signature=[REDACTED]",
		},
		"URL userinfo": {
			input: "unable to clone https://user:github-token@github.com/acme/repo.git",
			want:  "unable to clone https://[REDACTED]@github.com/acme/repo.git",
		},
		"authorization and cookie headers": {
			input: "Authorization: Bearer bearer-token\nCookie: session=cookie-secret\nContent-Type: application/json",
			want:  "Authorization: [REDACTED]\nCookie: [REDACTED]\nContent-Type: application/json",
		},
		"JSON credential": {
			input: `request body: {"token":"json-secret","name":"safe"}`,
			want:  `request body: {"token":"[REDACTED]","name":"safe"}`,
		},
		"bare assignment": {
			input: "command failed with TOKEN=bare-secret",
			want:  "command failed with TOKEN=[REDACTED]",
		},
		"prefixed environment variables": {
			input: "AWS_SECRET_ACCESS_KEY=aws-secret GITHUB_TOKEN=github-secret DATABASE_PASSWORD=db-secret",
			want:  "AWS_SECRET_ACCESS_KEY=[REDACTED] GITHUB_TOKEN=[REDACTED] DATABASE_PASSWORD=[REDACTED]",
		},
		"prefixed JSON credential": {
			input: `request body: {"github_token":"json-secret","name":"safe"}`,
			want:  `request body: {"github_token":"[REDACTED]","name":"safe"}`,
		},
		"quoted HCL assignments": {
			input: `password = "hcl-secret" and api_token='single-quoted-secret'`,
			want:  `password = "[REDACTED]" and api_token='[REDACTED]'`,
		},
		"HTML encoded separator": {
			input: "https://example.com?first=safe&amp;token=html-secret&amp;last=safe",
			want:  "https://example.com?first=safe&amp;token=[REDACTED]&amp;last=safe",
		},
		"benign URL": {
			input: "https://example.com?workspace_id=tfw123&operation=plan",
			want:  "https://example.com?workspace_id=tfw123&operation=plan",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := RedactDiagnosticSecrets(test.input); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestNew_RedactsDiagnosticSecretsFromAllPayloadFields(t *testing.T) {
	const querySecret = "tok_test_secret"
	const nestedSecret = "nested-secret"
	const hintSecret = "hint-secret"
	output := "GET https://api.example.com/state?workspace_id=tfw123&token=" + querySecret + "&operation=plan"

	d, err := New(sensitiveError{
		Output: output,
		Nested: map[string]any{
			"items":        []any{"Authorization: Bearer " + nestedSecret},
			"github_token": "plain-nested-secret",
		},
		hints: Hints{
			HintDocsURL: "https://docs.example.com/help?access_token=" + hintSecret,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name, value := range map[string]string{
		"message": d.Message,
		"section": d.Sections[0].Body,
		"data":    string(d.Data),
		"hint":    d.Hints[HintDocsURL],
	} {
		for _, secret := range []string{querySecret, nestedSecret, hintSecret} {
			if strings.Contains(value, secret) {
				t.Errorf("%s contains sensitive value %q: %q", name, secret, value)
			}
		}
	}
	if !strings.Contains(d.Message, "token="+redactedValue) || !strings.Contains(d.Message, "operation=plan") {
		t.Errorf("message was not redacted without losing safe query parameters: %q", d.Message)
	}
	if !strings.Contains(string(d.Data), `Authorization: [REDACTED]`) {
		t.Errorf("nested data was not redacted: %s", d.Data)
	}
	if !strings.Contains(string(d.Data), `"github_token":"[REDACTED]"`) {
		t.Errorf("credential-keyed nested data was not redacted: %s", d.Data)
	}
	if !strings.Contains(d.Hints[HintDocsURL], "access_token="+redactedValue) {
		t.Errorf("hint was not redacted: %q", d.Hints[HintDocsURL])
	}
}

func TestRedactJSONStrings_PreservesUnchangedPayload(t *testing.T) {
	data := []byte(`{"z":1,"a":"https://example.com?safe=value"}`)
	got, err := redactJSONStrings(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unchanged payload was rewritten: got %s, want %s", got, data)
	}
}

func TestRedactJSONStrings_PreservesNumbersWhenRedacting(t *testing.T) {
	data := []byte(`{"number":12345678901234567890,"github_token":"secret"}`)
	got, err := redactJSONStrings(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(got), `12345678901234567890`) {
		t.Fatalf("number changed during redaction: %s", got)
	}
	if strings.Contains(string(got), `"secret"`) || !strings.Contains(string(got), `"github_token":"[REDACTED]"`) {
		t.Fatalf("credential was not redacted: %s", got)
	}
}
