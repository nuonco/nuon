package orgs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// InterestsFlags collects the CLI flags shared by the `webhooks create` and
// `webhooks update` subcommands. Exactly one of (--all-events,
// --interests-json, --interests-file) may be set; --all-events is the implicit
// default when none is provided so that callers don't accidentally save an
// "empty" config that drops every event.
type InterestsFlags struct {
	AllEvents      bool
	InterestsJSON  string
	InterestsFile  string
	AllEventsIsSet bool
}

// Resolve turns the raw flag values into the JSON-shaped interests payload
// the API expects (mirrors internal/pkg/interests.Interests).
//
//   - --interests-file path : reads JSON from disk
//   - --interests-json '{…}': parses inline JSON
//   - --all-events / nothing: sends {"all_events":true}
//
// Server-side validation is authoritative; the CLI only ensures the bytes
// parse as JSON before sending. Mixing flags returns an error to keep CLI
// semantics unambiguous.
func (f InterestsFlags) Resolve() (any, error) {
	jsonProvided := strings.TrimSpace(f.InterestsJSON) != ""
	fileProvided := strings.TrimSpace(f.InterestsFile) != ""

	count := 0
	if jsonProvided {
		count++
	}
	if fileProvided {
		count++
	}
	if f.AllEventsIsSet {
		count++
	}
	if count > 1 {
		return nil, fmt.Errorf("only one of --all-events, --interests-json, --interests-file may be set")
	}

	switch {
	case fileProvided:
		raw, err := os.ReadFile(strings.TrimSpace(f.InterestsFile))
		if err != nil {
			return nil, fmt.Errorf("read --interests-file: %w", err)
		}
		return parseInterestsJSON(raw)

	case jsonProvided:
		return parseInterestsJSON([]byte(f.InterestsJSON))

	case f.AllEventsIsSet && !f.AllEvents:
		// Explicit --all-events=false with no per-resource config is almost
		// certainly user error — bail out rather than silently dropping
		// every event.
		return nil, fmt.Errorf("--all-events=false requires --interests-json or --interests-file with a per-resource config")

	default:
		return map[string]any{"all_events": true}, nil
	}
}

func parseInterestsJSON(raw []byte) (any, error) {
	var cfg any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse interests JSON: %w", err)
	}
	return cfg, nil
}
