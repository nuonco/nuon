package orgs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/services/orgs/subscriptiontui"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

// SubscriptionFlags collects the CLI flags shared by `webhooks create` and
// `webhooks update` describing the full webhook subscription — both the
// interests filter (which events fire) and the optional match predicate
// (which entities are in scope).
//
// Exactly one of (--subscription-json, --subscription-file) may be set.
// When neither is set, Resolve returns the implicit default: subscribe to
// every event in the org (interests = {"all_events": true}, match = nil).
type SubscriptionFlags struct {
	JSON string
	File string
}

// SubscriptionPayload is the parsed shape of the unified subscription file
// or inline JSON. Both fields are optional. The wire types are deliberately
// `any` so the CLI is agnostic to the server's internal Interests struct
// (which lives in services/ctl-api/internal/pkg/interests and is not part
// of the public SDK).
//
// Match is a *labels.SubscriptionMatch so we can run client-side Validate()
// before paying for an HTTP round-trip; a nil pointer (or omitted "match"
// key) preserves the legacy "match every event in the org" semantics.
type SubscriptionPayload struct {
	Interests any                       `json:"interests"`
	Match     *labels.SubscriptionMatch `json:"match"`
}

// Resolve turns the raw flag values into the parsed subscription payload
// the request builders pass into the SDK.
//
//   - --subscription-file path : reads JSON from disk
//   - --subscription-json '{…}': parses inline JSON
//   - neither                  : defaults to AllEvents + org-wide
//
// The JSON shape is `{"interests": {...}, "match": {...}}` — both keys
// optional. A missing "interests" defaults to {"all_events": true} so the
// caller can scope a webhook with `--subscription-json '{"match":{...}}'`
// without restating the events filter.
func (f SubscriptionFlags) Resolve() (SubscriptionPayload, error) {
	jsonProvided := strings.TrimSpace(f.JSON) != ""
	fileProvided := strings.TrimSpace(f.File) != ""

	if jsonProvided && fileProvided {
		return SubscriptionPayload{}, fmt.Errorf("only one of --subscription-json, --subscription-file may be set")
	}

	if !jsonProvided && !fileProvided {
		return SubscriptionPayload{
			Interests: map[string]any{"all_events": true},
		}, nil
	}

	var raw []byte
	if fileProvided {
		b, err := os.ReadFile(strings.TrimSpace(f.File))
		if err != nil {
			return SubscriptionPayload{}, fmt.Errorf("read --subscription-file: %w", err)
		}
		raw = b
	} else {
		raw = []byte(f.JSON)
	}

	var payload SubscriptionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return SubscriptionPayload{}, fmt.Errorf("parse subscription JSON: %w", err)
	}

	if payload.Interests == nil {
		payload.Interests = map[string]any{"all_events": true}
	}

	if payload.Match != nil {
		if err := payload.Match.Validate(); err != nil {
			return SubscriptionPayload{}, fmt.Errorf("invalid match: %w", err)
		}
	}

	return payload, nil
}

// resolveSubscription drops into the interactive picker when neither
// --subscription-json nor --subscription-file was supplied AND the session
// is interactive. Otherwise it falls through to the non-interactive
// Resolve path (validated JSON if a flag was set, implicit AllEvents
// default if not). Keeping the TUI gated on cfg.Interactive matches the
// CLI's TTY conventions — `nuon … | cat`, `CI=true`, and `NUON_NO_TTY=true`
// all preserve the existing scriptable default behaviour.
//
// The picker itself lives in the subscriptiontui subpackage so the heavy
// huh / lipgloss / bubbletea import graph is only linked into command
// paths that can actually invoke it. The non-TUI orgs commands (delete,
// list, get, …) never reach this function.
//
// ctx + api are threaded through because the "specific" match mode runs
// data-driven entity pickers (apps → components, apps → actions,
// org-wide installs) that hit the SDK between huh forms.
func resolveSubscription(ctx context.Context, api nuon.Client, interactive bool, f SubscriptionFlags) (SubscriptionPayload, error) {
	jsonProvided := strings.TrimSpace(f.JSON) != ""
	fileProvided := strings.TrimSpace(f.File) != ""
	if interactive && !jsonProvided && !fileProvided {
		interestsCfg, match, err := subscriptiontui.Run(ctx, api)
		if err != nil {
			return SubscriptionPayload{}, err
		}
		return SubscriptionPayload{Interests: interestsCfg, Match: match}, nil
	}
	return f.Resolve()
}
