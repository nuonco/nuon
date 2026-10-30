package azure

import "testing"

func TestConfiguredRunnerAudience(t *testing.T) {
	const audience = "api://11111111-1111-1111-1111-111111111111"
	if got := ConfiguredRunnerAudience(audience); got != audience {
		t.Fatalf("audience = %q", got)
	}
	for _, bad := range []string{"", ManagementAudience, "api://not-a-guid", "https://example.com"} {
		if got := ConfiguredRunnerAudience(bad); got != "" {
			t.Fatalf("configured %q accepted as %q", bad, got)
		}
	}
}

func TestAuthAudiences(t *testing.T) {
	got := AuthAudiences("")
	if len(got) != 1 || got[0] != ManagementAudience {
		t.Fatalf("empty config = %v", got)
	}

	const audience = "api://11111111-1111-1111-1111-111111111111"
	got = AuthAudiences(audience)
	if len(got) != 2 || got[0] != ManagementAudience || got[1] != audience {
		t.Fatalf("configured = %v", got)
	}

	got = AuthAudiences("nope")
	if len(got) != 1 || got[0] != ManagementAudience {
		t.Fatalf("invalid config = %v", got)
	}
}

func TestRunnerAuthResource(t *testing.T) {
	got, ok := RunnerAuthResource("")
	if !ok || got != ManagementAudience {
		t.Fatalf("empty = %q %v", got, ok)
	}

	const audience = "api://11111111-1111-1111-1111-111111111111"
	got, ok = RunnerAuthResource(audience)
	if !ok || got != audience {
		t.Fatalf("app audience = %q %v", got, ok)
	}

	if _, ok := RunnerAuthResource("https://graph.microsoft.com"); ok {
		t.Fatal("accepted a non-runner audience")
	}
}
