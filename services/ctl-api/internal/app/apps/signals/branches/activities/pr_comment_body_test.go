package activities

import (
	"strings"
	"testing"
)

func TestBuildPRCommentBodyIncludesInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSuccess,
		InstallImpact: []InstallGroupImpact{
			{
				GroupName: "canary",
				Installs: []InstallImpact{
					{InstallID: "insta", InstallName: "canary-1", Added: 1, Changed: 2, StackChanged: true},
				},
			},
			{
				GroupName: "prod",
				Installs: []InstallImpact{
					{InstallID: "instb", InstallName: "prod-1", Unchanged: 5},
				},
			},
		},
	})

	for _, want := range []string{
		"Install Impact — 2 install(s)",
		"nothing was applied",
		"canary",
		"canary-1",
		"prod-1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment body missing %q\n%s", want, body)
		}
	}
}

func TestBuildPRCommentBodyOmitsEmptyInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSuccess,
	})

	if strings.Contains(body, "Install Impact") {
		t.Errorf("expected no install impact section when there are no installs\n%s", body)
	}
}

// A no-changes preview short-circuits before any install is resolved, so the
// impact section would be misleading there.
func TestBuildPRCommentBodySkippedHasNoInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSkipped,
		InstallImpact: []InstallGroupImpact{
			{GroupName: "prod", Installs: []InstallImpact{{InstallID: "instb", InstallName: "prod-1"}}},
		},
	})

	if strings.Contains(body, "Install Impact") {
		t.Errorf("skipped preview should not render install impact\n%s", body)
	}
}
