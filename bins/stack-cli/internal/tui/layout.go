package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// narrowBreakpoint is the body width below which the two-column layout
// collapses to a single stacked column (main on top, detail beneath a rule).
// Above it, main and detail share the width 50/50 with a vertical rule between.
const narrowBreakpoint = 96

// twoColumn renders the standardized step body. main fills the left column and
// detail the right; each is asked to render at its own column width so it can
// wrap and scroll to fit. An empty detail collapses to a single full-width
// column at any size.
func twoColumn(w, h int, main, detail func(width, height int) string) string {
	if w < narrowBreakpoint {
		m := main(w, h)
		d := detail(w, h)
		if strings.TrimSpace(d) == "" {
			return m
		}
		rule := ruleStyle.Render(strings.Repeat("─", w))
		return strings.Join([]string{m, "", rule, "", d}, "\n")
	}

	const sep = 3 // " │ "
	leftW := (w - sep) / 3
	rightW := w - sep - leftW

	left := lipgloss.NewStyle().Width(leftW).Render(main(leftW, h))
	right := lipgloss.NewStyle().Width(rightW).Render(detail(rightW, h))

	divLines := make([]string, h)
	for i := range divLines {
		divLines[i] = dividerStyle.Render("│")
	}
	divider := strings.Join(divLines, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", divider, " ", right)
}
