package activities

import (
	"fmt"
	"strings"
	"time"
)

type PRCommentStatus string

const (
	PRCommentStatusPending PRCommentStatus = "pending"
	PRCommentStatusSuccess PRCommentStatus = "success"
	PRCommentStatusFailed  PRCommentStatus = "failed"
	PRCommentStatusSkipped PRCommentStatus = "skipped"
)

type PRCommentParams struct {
	AppName      string
	RunID        string
	RunURL       string
	Status       PRCommentStatus
	Diff         *ComputeAppConfigDiffOutput
	ErrorMessage string
}

func BuildPRCommentBody(p *PRCommentParams) string {
	var b strings.Builder

	appName := p.AppName
	if appName == "" {
		appName = "App"
	}

	b.WriteString(fmt.Sprintf("## Nuon Preview \u2014 %s\n\n", appName))

	switch p.Status {
	case PRCommentStatusPending:
		b.WriteString("**Status**: \u23f3 In Progress\n\n")
	case PRCommentStatusSuccess:
		b.WriteString("**Status**: \u2705 Complete\n\n")
	case PRCommentStatusFailed:
		b.WriteString("**Status**: \u274c Failed\n\n")
	case PRCommentStatusSkipped:
		b.WriteString("**Status**: \u2298 No Changes\n\n")
	}

	if p.Status == PRCommentStatusSkipped {
		b.WriteString("No changes to `nuon.toml` detected in this PR. Preview skipped.\n")
	} else if p.Status == PRCommentStatusFailed && p.ErrorMessage != "" {
		if p.Diff != nil {
			writeDiffSection(&b, p.Diff)
		}
		b.WriteString("### Error\n\n")
		b.WriteString(fmt.Sprintf("```\n%s\n```\n", p.ErrorMessage))
	} else if p.Status == PRCommentStatusPending {
		if p.Diff != nil {
			writeDiffSection(&b, p.Diff)
			b.WriteString("\u23f3 Building components...\n")
		} else {
			b.WriteString("\u23f3 Parsing config...\n")
		}
	} else if p.Diff != nil {
		writeDiffSection(&b, p.Diff)
	}

	b.WriteString("\n---\n")
	runLink := fmt.Sprintf("`%s`", p.RunID)
	if p.RunURL != "" {
		runLink = fmt.Sprintf("[View Run](%s)", p.RunURL)
	}
	b.WriteString(fmt.Sprintf("*[Nuon](https://nuon.co) \u2022 %s \u2022 Updated: %s*\n",
		runLink, time.Now().UTC().Format("Jan 2, 2006 3:04 PM UTC")))

	return b.String()
}

func writeDiffSection(b *strings.Builder, diff *ComputeAppConfigDiffOutput) {
	b.WriteString(fmt.Sprintf("### Config Changes \u2014 `%s`\n\n", diff.ConfigFile))

	if len(diff.Sections) == 0 {
		b.WriteString("No sections changed.\n\n")
		return
	}

	b.WriteString("| Section | Added | Changed | Removed |\n")
	b.WriteString("|---------|-------|---------|--------|\n")
	for _, s := range diff.Sections {
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			s.Name,
			formatCount(s.Additions, "+"),
			formatCount(s.Changed, ""),
			formatCount(s.Removals, ""),
		))
	}
	b.WriteString("\n")

	for _, s := range diff.Sections {
		if len(s.Entries) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("<details><summary>%s (%d)</summary>\n\n", s.Name, len(s.Entries)))
		for _, e := range s.Entries {
			prefix := "\u2022"
			switch e.Op {
			case "add":
				prefix = "**+**"
			case "remove":
				prefix = "**-**"
			case "change":
				prefix = "**~**"
			}
			desc := ""
			if e.Description != "" {
				desc = fmt.Sprintf(" \u2014 %s", e.Description)
			}
			b.WriteString(fmt.Sprintf("- %s `%s`%s\n", prefix, e.Name, desc))
		}
		b.WriteString("\n</details>\n\n")
	}
}

func formatCount(n int, prefix string) string {
	if n == 0 {
		return "0"
	}
	return fmt.Sprintf("%s%d", prefix, n)
}
