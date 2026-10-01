package service

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const mcpCompositeErrorBodyMaxLen = 4000

type mcpCompositeError struct {
	Type     string                     `json:"type"`
	Severity string                     `json:"severity"`
	Message  string                     `json:"message"`
	Sections []mcpCompositeErrorSection `json:"sections,omitempty"`
	Hints    *mcpCompositeErrorHints    `json:"hints,omitempty"`
}

type mcpCompositeErrorSection struct {
	Heading string `json:"heading"`
	Kind    string `json:"kind,omitempty"`
	Body    string `json:"body"`
}

type mcpCompositeErrorHints struct {
	SkipAutoRetry bool   `json:"skip_auto_retry,omitempty"`
	Terminal      bool   `json:"terminal,omitempty"`
	DocsURL       string `json:"docs_url,omitempty"`
}

func mcpCompositeErrorFrom(data *compositeerrors.CompositeErrorData) *mcpCompositeError {
	if data == nil || (data.Type == "" && data.Message == "") {
		return nil
	}

	out := &mcpCompositeError{
		Type:     string(data.Type),
		Severity: string(data.Severity),
		Message:  data.Message,
	}
	for _, section := range data.Sections {
		out.Sections = append(out.Sections, mcpCompositeErrorSection{
			Heading: section.Heading,
			Kind:    string(section.Kind),
			Body:    clipRunes(section.Body, mcpCompositeErrorBodyMaxLen),
		})
	}

	hints := &mcpCompositeErrorHints{
		SkipAutoRetry: data.Hints.SkipAutoRetry(),
		Terminal:      data.Hints.Terminal(),
		DocsURL:       data.Hints.DocsURL(),
	}
	if hints.SkipAutoRetry || hints.Terminal || hints.DocsURL != "" {
		out.Hints = hints
	}
	return out
}

func clipRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
