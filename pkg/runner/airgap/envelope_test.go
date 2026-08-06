package airgap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvelope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "envelope.json")
	contents := `{"version":"v0","org_id":"org","app_id":"app","install_id":"install","steps":[{"id":"one","name":"one","job_type":"noop-deploy","job_operation":"exec","job_group":"deploy","composite_plan":{}},{"id":"two","name":"two","job_type":"noop-deploy","job_operation":"exec","job_group":"deploy","depends_on":["one"],"composite_plan":{}}]}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	envelope, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(envelope.Steps) != 2 || !json.Valid(envelope.Steps[0].CompositePlan) {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
}

func TestParseEnvelope(t *testing.T) {
	envelope, err := Parse([]byte(`{"version":"v0","org_id":"org","app_id":"app","steps":[{"id":"one","name":"one","job_type":"noop-deploy","job_operation":"exec","job_group":"deploy","composite_plan":{}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if envelope.OrgID != "org" || len(envelope.Steps) != 1 {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}

	if _, err := Parse([]byte(`{"version":"v1","steps":[]}`)); err == nil || !strings.Contains(err.Error(), "unsupported envelope version") {
		t.Fatalf("expected version error, got %v", err)
	}
	if _, err := Parse([]byte(`not json`)); err == nil || !strings.Contains(err.Error(), "decode envelope") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestEnvelopeCycle(t *testing.T) {
	envelope := &Envelope{Version: "v0", Steps: []Step{
		{ID: "one", JobType: "noop-deploy", JobGroup: "deploy", DependsOn: []string{"two"}},
		{ID: "two", JobType: "noop-deploy", JobGroup: "deploy", DependsOn: []string{"one"}},
	}}
	err := envelope.Validate()
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}
