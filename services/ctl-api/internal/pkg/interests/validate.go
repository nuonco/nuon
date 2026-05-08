package interests

import "fmt"

// Validate checks that resource keys + ops + outcome on an Interests config
// match the canonical taxonomy declared in this package. Empty configs and
// AllEvents=true are both valid; the picker UI is the place where shape is
// enforced visually, but the API still refuses outright garbage.
//
// Used by both the slack channel subscription create handler and the webhook
// create/update handlers. Lifted from internal/app/slack/service so both
// surfaces share one implementation.
func Validate(in Interests) error {
	if in.AllEvents || len(in.Resources) == 0 {
		return nil
	}
	for kind, cfg := range in.Resources {
		validOps, ok := SubOps[kind]
		if !ok {
			return fmt.Errorf("invalid interests: unknown resource %q", kind)
		}
		for _, op := range cfg.Ops {
			found := false
			for _, valid := range validOps {
				if valid == op {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid interests: unknown op %q for resource %q", op, kind)
			}
		}
		switch cfg.Outcome {
		// OutcomeNone is a legitimate value (mute lifecycle entirely;
		// approvals / drift still fire). It's documented in types.go
		// and explicitly handled in match.go, so the validator must
		// accept it too.
		case "", OutcomeAll, OutcomeCompletion, OutcomeFailures, OutcomeNone:
		default:
			return fmt.Errorf("invalid interests: unknown outcome %q for resource %q", cfg.Outcome, kind)
		}
	}
	return nil
}
