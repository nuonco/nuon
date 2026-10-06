package agentclient

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestTouchCreatesThenUpdates(t *testing.T) {
	dir := t.TempDir()
	first := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	created, err := Touch(dir, Record{
		Agent:      Cursor,
		CLIVersion: "1.0.0",
		AppID:      "app-1",
	}, first)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected first write to create the file")
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %o, want 700", info.Mode().Perm())
	}
	fileInfo, err := os.Stat(filepath.Join(dir, "cursor.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %o, want 600", fileInfo.Mode().Perm())
	}

	later := first.Add(time.Hour)
	created, err = Touch(dir, Record{
		Agent:      Cursor,
		CLIVersion: "1.1.0",
		AppID:      "app-2",
	}, later)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected update to keep the existing file")
	}

	body, err := os.ReadFile(filepath.Join(dir, "cursor.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var rec Record
	if err := yaml.Unmarshal(body, &rec); err != nil {
		t.Fatal(err)
	}
	if !rec.FirstSeen.Equal(first) {
		t.Fatalf("first_seen = %s, want %s", rec.FirstSeen, first)
	}
	if !rec.LastSeen.Equal(later) {
		t.Fatalf("last_seen = %s, want %s", rec.LastSeen, later)
	}
	if rec.CLIVersion != "1.1.0" || rec.AppID != "app-2" || rec.Agent != Cursor {
		t.Fatalf("record = %+v", rec)
	}
}

func TestFirstRunGuide(t *testing.T) {
	guide := FirstRunGuide(Claude)
	for _, want := range []string{
		"Nuon agent setup (claude)",
		"https://docs.nuon.co/guides/agents",
		"nuon agents help",
		"https://app.nuon.co",
		"nuon auth login",
		"nuon orgs select",
		"nuon apps select",
		"nuon installs list",
		"nuon agents mcp",
	} {
		if !strings.Contains(guide, want) {
			t.Errorf("guide missing %q\n%s", want, guide)
		}
	}
}

func TestUserAgent(t *testing.T) {
	got := UserAgent("1.2.3", Cursor)
	if got != "nuon-cli/1.2.3 (cursor)" {
		t.Fatalf("UserAgent = %q", got)
	}
	if !strings.Contains(got, "nuon-cli") {
		t.Fatal("User-Agent must stay recognizable as the CLI")
	}
}
