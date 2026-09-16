package activities

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const previewCommentCollapsedMarker = "<!-- nuon-preview-comment-collapsed -->"

// Every preview report carries this marker so a later run can recognise its own
// earlier comments on the PR without relying on comment order or authorship.
var previewCommentMarkerRe = regexp.MustCompile(`<!-- nuon-preview-comment name="([^"]*)"(?: run="([^"]*)")? -->`)

// PreviewCommentMarker is the machine-readable header of a preview PR comment.
type PreviewCommentMarker struct {
	Name  string
	RunID string
}

func previewCommentMarkerLine(name, runID string) string {
	if runID == "" {
		return fmt.Sprintf(`<!-- nuon-preview-comment name=%q -->`, sanitizeMarkerValue(name))
	}
	return fmt.Sprintf(`<!-- nuon-preview-comment name=%q run=%q -->`,
		sanitizeMarkerValue(name), sanitizeMarkerValue(runID))
}

func sanitizeMarkerValue(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '"', '<', '>', '\n', '\r':
			return -1
		}
		return r
	}, value)
}

// ParsePreviewCommentMarker reads the marker out of a PR comment body,
// reporting false for anything Nuon did not write.
func ParsePreviewCommentMarker(body string) (PreviewCommentMarker, bool) {
	match := previewCommentMarkerRe.FindStringSubmatch(body)
	if match == nil {
		return parseLegacyPreviewComment(body)
	}
	return PreviewCommentMarker{Name: match[1], RunID: match[2]}, true
}

// Reports posted before markers existed are recognised from their rendered
// heading, so PRs that already carry a stack of them still get collapsed. Both
// the heading and a report-only line are required: a human quoting the heading
// must not have their comment rewritten.
var (
	legacyPreviewTitleRe = regexp.MustCompile(`(?m)^## Nuon Preview \x{2014} (.+)$`)
	legacyPreviewRunRe   = regexp.MustCompile("(?m)^Preview run: `([^`]+)`")
)

func parseLegacyPreviewComment(body string) (PreviewCommentMarker, bool) {
	match := legacyPreviewTitleRe.FindStringSubmatch(body)
	if match == nil {
		return PreviewCommentMarker{}, false
	}
	if !strings.Contains(body, "Preview run:") && !strings.Contains(body, "View preview run") {
		return PreviewCommentMarker{}, false
	}

	marker := PreviewCommentMarker{Name: stripPreviewModeLabel(strings.TrimSpace(match[1]))}
	if run := legacyPreviewRunRe.FindStringSubmatch(body); run != nil {
		marker.RunID = run[1]
	}
	return marker, true
}

func stripPreviewModeLabel(title string) string {
	for _, mode := range []app.AppBranchRunPreviewMode{
		app.AppBranchRunPreviewModeBuildOnly,
		app.AppBranchRunPreviewModePlanOnly,
		app.AppBranchRunPreviewModeApply,
	} {
		label := mode.Label()
		if label == "" {
			continue
		}
		if trimmed := strings.TrimSuffix(title, " ("+label+")"); trimmed != title {
			return trimmed
		}
	}
	return title
}

// IsCollapsedPreviewComment reports whether the body has already been folded
// into its collapsed form, so repeated collapse passes are no-ops.
func IsCollapsedPreviewComment(body string) bool {
	return strings.Contains(body, previewCommentCollapsedMarker)
}

// CollapsePreviewCommentBody rewrites a preview report as a collapsed
// <details> block, keeping the full report one click away. It returns false
// when the body is not a preview report or is already collapsed.
func CollapsePreviewCommentBody(body string) (string, bool) {
	marker, ok := ParsePreviewCommentMarker(body)
	if !ok || IsCollapsedPreviewComment(body) {
		return body, false
	}

	report := strings.TrimSpace(stripPreviewCommentMarkers(body))
	if report == "" {
		return body, false
	}

	var b strings.Builder
	b.WriteString(previewCommentMarkerLine(marker.Name, marker.RunID) + "\n")
	b.WriteString(previewCommentCollapsedMarker + "\n\n")
	b.WriteString("<details>\n")
	b.WriteString(fmt.Sprintf("<summary>%s</summary>\n\n", collapsedPreviewSummary(marker)))
	b.WriteString(report + "\n\n")
	b.WriteString("</details>\n")
	return b.String(), true
}

func collapsedPreviewSummary(marker PreviewCommentMarker) string {
	summary := fmt.Sprintf("\U0001f5c2\ufe0f Outdated Nuon Preview \u2014 <code>%s</code>", marker.Name)
	if marker.RunID != "" {
		summary += fmt.Sprintf(" (run <code>%s</code>)", marker.RunID)
	}
	return summary
}

func stripPreviewCommentMarkers(body string) string {
	lines := strings.Split(body, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == previewCommentCollapsedMarker || previewCommentMarkerRe.MatchString(trimmed) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
