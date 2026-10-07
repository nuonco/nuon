package activities

import "testing"

func TestBuildChangeReason(t *testing.T) {
	if got := buildChangeReason(true, "old", "new"); got != ChangeReasonSourceAndConfig {
		t.Fatalf("both: got %s", got)
	}
	if got := buildChangeReason(true, "same", "same"); got != ChangeReasonSourceChanged {
		t.Fatalf("source: got %s", got)
	}
	if got := buildChangeReason(false, "old", "new"); got != ChangeReasonConfigChanged {
		t.Fatalf("config: got %s", got)
	}
}
