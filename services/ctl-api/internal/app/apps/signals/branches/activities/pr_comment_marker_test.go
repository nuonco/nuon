package activities

import (
	"strings"
	"testing"

	"github.com/google/go-github/v50/github"
)

func TestBuildPRCommentBodyCarriesPreviewMarker(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:    "acme",
		AppName:    "payments",
		BranchName: "production",
		RunID:      "abrq7fplr1up5atx5zpxotbabm",
		Status:     PRCommentStatusSuccess,
	})

	marker, ok := ParsePreviewCommentMarker(body)
	if !ok {
		t.Fatalf("comment body has no preview marker\n%s", body)
	}
	if marker.Name != "acme/payments/production" {
		t.Errorf("unexpected preview name %q", marker.Name)
	}
	if marker.RunID != "abrq7fplr1up5atx5zpxotbabm" {
		t.Errorf("unexpected run id %q", marker.RunID)
	}
	if IsCollapsedPreviewComment(body) {
		t.Error("a freshly built comment should not be collapsed")
	}
}

func TestParsePreviewCommentMarkerRejectsForeignComments(t *testing.T) {
	for _, body := range []string{
		"",
		"LGTM",
		"## Nuon Preview — acme/payments/production",
		"quoting the bot:\n\n> ## Nuon Preview — acme/payments/production\n",
	} {
		if _, ok := ParsePreviewCommentMarker(body); ok {
			t.Errorf("expected no marker in %q", body)
		}
	}
}

// Comments posted before markers existed are matched on their heading so an
// already-noisy PR collapses on the next run.
func TestParsePreviewCommentMarkerFallsBackToHeading(t *testing.T) {
	body := "## Nuon Preview — acme/payments/production (apply)\n\n" +
		"Preview run: `abrq7fplr1up5atx5zpxotbabm`\n\n" +
		"**Status**: ✅ Complete\n"

	marker, ok := ParsePreviewCommentMarker(body)
	if !ok {
		t.Fatalf("expected legacy preview comment to be recognised\n%s", body)
	}
	if marker.Name != "acme/payments/production" {
		t.Errorf("unexpected preview name %q", marker.Name)
	}
	if marker.RunID != "abrq7fplr1up5atx5zpxotbabm" {
		t.Errorf("unexpected run id %q", marker.RunID)
	}

	collapsed, ok := CollapsePreviewCommentBody(body)
	if !ok {
		t.Fatal("expected legacy preview comment to collapse")
	}
	if !IsCollapsedPreviewComment(collapsed) || !strings.Contains(collapsed, "<details>") {
		t.Errorf("legacy comment was not collapsed\n%s", collapsed)
	}
}

func TestCollapsePreviewCommentBody(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:    "acme",
		AppName:    "payments",
		BranchName: "production",
		RunID:      "abrq7fplr1up5atx5zpxotbabm",
		Status:     PRCommentStatusSuccess,
	})

	collapsed, ok := CollapsePreviewCommentBody(body)
	if !ok {
		t.Fatal("expected the preview comment to collapse")
	}
	if !IsCollapsedPreviewComment(collapsed) {
		t.Errorf("collapsed body is missing its marker\n%s", collapsed)
	}
	if !strings.Contains(collapsed, "<summary>") || !strings.Contains(collapsed, "</details>") {
		t.Errorf("collapsed body is not a details block\n%s", collapsed)
	}
	if !strings.Contains(collapsed, "Outdated Nuon Preview") ||
		!strings.Contains(collapsed, "acme/payments/production") ||
		!strings.Contains(collapsed, "abrq7fplr1up5atx5zpxotbabm") {
		t.Errorf("collapsed summary does not identify the preview run\n%s", collapsed)
	}
	if !strings.Contains(collapsed, "## Nuon Preview") {
		t.Errorf("collapsed body dropped the original report\n%s", collapsed)
	}

	marker, ok := ParsePreviewCommentMarker(collapsed)
	if !ok || marker.Name != "acme/payments/production" {
		t.Errorf("collapsed body lost its preview marker: %+v", marker)
	}
	if strings.Count(collapsed, "nuon-preview-comment name=") != 1 {
		t.Errorf("collapsed body duplicated the preview marker\n%s", collapsed)
	}

	if _, ok := CollapsePreviewCommentBody(collapsed); ok {
		t.Error("collapsing an already collapsed comment should be a no-op")
	}
}

func TestCollapsePreviewCommentBodyIgnoresForeignComments(t *testing.T) {
	if _, ok := CollapsePreviewCommentBody("looks good to me"); ok {
		t.Error("a human comment must never be rewritten")
	}
}

// Reports for different previews coexist on one PR, so only same-named
// comments are candidates for collapsing.
func TestPreviewCommentNameSeparatesApps(t *testing.T) {
	payments := PreviewCommentName("acme", "payments", "production")
	billing := PreviewCommentName("acme", "billing", "production")

	if payments == billing {
		t.Fatalf("expected distinct preview names, got %q", payments)
	}
	if got := PreviewCommentName("", "", ""); got != "App" {
		t.Errorf("unexpected fallback preview name %q", got)
	}
}

func TestNewestLivePreviewCommentIDSkipsCollapsed(t *testing.T) {
	live := BuildPRCommentBody(&PRCommentParams{AppName: "payments", RunID: "abr2", Status: PRCommentStatusSuccess})
	collapsed, _ := CollapsePreviewCommentBody(
		BuildPRCommentBody(&PRCommentParams{AppName: "payments", RunID: "abr1", Status: PRCommentStatusSuccess}),
	)

	comments := []*github.IssueComment{
		{ID: github.Int64(1), Body: github.String(collapsed)},
		{ID: github.Int64(2), Body: github.String(live)},
		{ID: github.Int64(3), Body: github.String(collapsed)},
	}

	if got := newestLivePreviewCommentID(comments); got != 2 {
		t.Errorf("expected to reuse comment 2, got %d", got)
	}
	if got := newestLivePreviewCommentID(nil); got != 0 {
		t.Errorf("expected no comment to reuse, got %d", got)
	}
}
