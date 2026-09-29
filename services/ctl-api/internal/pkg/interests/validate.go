package interests

import "fmt"

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
		case "", OutcomeAll, OutcomeCompletion, OutcomeFailures, OutcomeNone:
		default:
			return fmt.Errorf("invalid interests: unknown outcome %q for resource %q", cfg.Outcome, kind)
		}
	}
	return nil
}
