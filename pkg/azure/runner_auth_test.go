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

func TestExpectedRunnerAudience(t *testing.T) {
	const audience = "api://11111111-1111-1111-1111-111111111111"
	const other = "api://22222222-2222-2222-2222-222222222222"
	cases := []struct {
		configured, presented, want string
	}{
		{"", ManagementAudience, ManagementAudience},
		{"", audience, ManagementAudience},
		{audience, audience, audience},
		{audience, ManagementAudience, ManagementAudience},
		{audience, other, ManagementAudience},
		{audience, "", ManagementAudience},
		{"nope", "nope", ManagementAudience},
	}
	for _, c := range cases {
		if got := ExpectedRunnerAudience(c.configured, c.presented); got != c.want {
			t.Fatalf("ExpectedRunnerAudience(%q, %q) = %q, want %q", c.configured, c.presented, got, c.want)
		}
	}
}

func TestRunnerAuthAppID(t *testing.T) {
	const audience = "api://11111111-1111-1111-1111-111111111111"
	if got := RunnerAuthAppID(audience); got != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("app id = %q", got)
	}
	if got := RunnerAuthAppID(ManagementAudience); got != "" {
		t.Fatalf("arm audience produced app id %q", got)
	}
	if got := RunnerAuthAppID(""); got != "" {
		t.Fatalf("empty audience produced app id %q", got)
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
