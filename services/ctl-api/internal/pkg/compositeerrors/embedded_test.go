package compositeerrors

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/redact"
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
	if !strings.Contains(d.Message, "token="+redact.Placeholder) || !strings.Contains(d.Message, "operation=plan") {
		t.Errorf("message was not redacted without losing safe query parameters: %q", d.Message)
	}
	if !strings.Contains(string(d.Data), `Authorization: [REDACTED]`) {
		t.Errorf("nested data was not redacted: %s", d.Data)
	}
	if !strings.Contains(string(d.Data), `"github_token":"[REDACTED]"`) {
		t.Errorf("credential-keyed nested data was not redacted: %s", d.Data)
	}
	if !strings.Contains(d.Hints[HintDocsURL], "access_token="+redact.Placeholder) {
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
