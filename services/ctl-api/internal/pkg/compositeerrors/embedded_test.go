package compositeerrors

import "testing"

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
