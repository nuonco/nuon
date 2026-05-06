package unknown_error

import (
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
)

func init() {
	catalog.RegisterType(Type, func() composite_error.CompositeError {
		return &Error{}
	})
}

// Build constructs an unknown_error ParseResult from raw input. The pipeline
// uses this as its safety-net builder when no parser matches.
//
// Lives on the type package (rather than the pipeline package) so the rules
// for "what does an unknown error look like" stay alongside the type itself.
func Build(in composite_error.ParseInput) composite_error.ParseResult {
	msg := firstLine(string(in.Raw))
	if msg == "" && in.GoErr != nil {
		msg = in.GoErr.Error()
	}

	e := &Error{Message: msg}
	if in.ExitCode != 0 {
		ec := in.ExitCode
		e.ExitCode = &ec
	}

	src := composite_error.Source{
		ParserName:    "unknown_error.builder",
		ParserVersion: "1",
		Snippet:       composite_error.CapSnippet(string(in.Raw)),
	}
	if in.ExitCode != 0 {
		ec := in.ExitCode
		src.ExitCode = &ec
	}
	if in.GoErr != nil {
		src.GoError = in.GoErr.Error()
	}

	return composite_error.ParseResult{
		Matched: true,
		Error:   e,
		Source:  src,
	}
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
