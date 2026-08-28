package customermanaged

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type inputReference struct {
	index int
	name  string
	value string
}

func rewriteInputPlaceholders(ctx context.Context, db *gorm.DB, envelope *customermanaged.Envelope) error {
	if len(envelope.Inputs) == 0 {
		return nil
	}

	var installInputs app.InstallInputs
	err := db.WithContext(ctx).
		Where(app.InstallInputs{InstallID: envelope.InstallID}).
		Order("created_at DESC").
		First(&installInputs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("load latest inputs for reference install %s: %w", envelope.InstallID, err)
	}

	planCount := len(envelope.Steps) + len(envelope.Actions) + len(envelope.Drift)
	plans := make([]any, 0, planCount)
	for i := range envelope.Steps {
		var plan any
		if err := json.Unmarshal(envelope.Steps[i].CompositePlan, &plan); err != nil {
			return fmt.Errorf("decode composite plan for step %s: %w", envelope.Steps[i].ID, err)
		}
		plans = append(plans, plan)
	}
	for i := range envelope.Actions {
		var plan any
		if err := json.Unmarshal(envelope.Actions[i].CompositePlan, &plan); err != nil {
			return fmt.Errorf("decode composite plan for action %s: %w", envelope.Actions[i].ID, err)
		}
		plans = append(plans, plan)
	}
	for i := range envelope.Drift {
		var plan any
		if err := json.Unmarshal(envelope.Drift[i].CompositePlan, &plan); err != nil {
			return fmt.Errorf("decode composite plan for drift template %s: %w", envelope.Drift[i].ID, err)
		}
		plans = append(plans, plan)
	}

	references := make([]inputReference, 0, len(envelope.Inputs))
	for i := range envelope.Inputs {
		spec := &envelope.Inputs[i]
		value := spec.Default
		if installValue, ok := installInputs.Values[spec.Name]; ok && installValue != nil {
			value = *installValue
		}
		if value == "" {
			continue
		}
		if spec.Secret {
			if len(value) >= 6 && containsInputValue(plans, value) {
				return fmt.Errorf("input %q is a secret and its value is baked into the exported plans; portable bundles cannot ship or accept secrets", spec.Name)
			}
			continue
		}
		if !rewriteEligible(value) {
			continue
		}
		references = append(references, inputReference{index: i, name: spec.Name, value: value})
	}

	seen := make(map[string]string, len(references))
	for _, reference := range references {
		if previous, ok := seen[reference.value]; ok {
			return fmt.Errorf("inputs %q and %q have the same reference value; portable bundle inputs must be disambiguated", previous, reference.name)
		}
		seen[reference.value] = reference.name
	}
	sort.Slice(references, func(i, j int) bool {
		if len(references[i].value) != len(references[j].value) {
			return len(references[i].value) > len(references[j].value)
		}
		return references[i].name < references[j].name
	})

	for _, reference := range references {
		placeholder := customermanaged.InputPlaceholder(reference.name)
		for _, plan := range plans {
			if rewriteInputValue(plan, reference.value, placeholder) {
				envelope.Inputs[reference.index].Bindable = true
			}
		}
	}

	for i := range envelope.Steps {
		encoded, err := json.Marshal(plans[i])
		if err != nil {
			return fmt.Errorf("encode composite plan for step %s: %w", envelope.Steps[i].ID, err)
		}
		envelope.Steps[i].CompositePlan = encoded
	}
	offset := len(envelope.Steps)
	for i := range envelope.Actions {
		encoded, err := json.Marshal(plans[offset+i])
		if err != nil {
			return fmt.Errorf("encode composite plan for action %s: %w", envelope.Actions[i].ID, err)
		}
		envelope.Actions[i].CompositePlan = encoded
	}
	offset += len(envelope.Actions)
	for i := range envelope.Drift {
		encoded, err := json.Marshal(plans[offset+i])
		if err != nil {
			return fmt.Errorf("encode composite plan for drift template %s: %w", envelope.Drift[i].ID, err)
		}
		envelope.Drift[i].CompositePlan = encoded
	}
	return nil
}

func rewriteEligible(value string) bool {
	return len(value) >= 3 && value != "true" && value != "false"
}

func containsInputValue(nodes []any, value string) bool {
	for _, node := range nodes {
		if walkInputStrings(node, func(s string) (string, bool) {
			return s, strings.Contains(s, value)
		}) {
			return true
		}
	}
	return false
}

func rewriteInputValue(node any, value, placeholder string) bool {
	return walkInputStrings(node, func(s string) (string, bool) {
		if len(value) >= 6 {
			rewritten := strings.ReplaceAll(s, value, placeholder)
			return rewritten, rewritten != s
		}
		if s == value {
			return placeholder, true
		}
		if !strings.Contains(s, ",") {
			return s, false
		}
		parts := strings.Split(s, ",")
		changed := false
		for i := range parts {
			if parts[i] == value {
				parts[i] = placeholder
				changed = true
			}
		}
		return strings.Join(parts, ","), changed
	})
}

func walkInputStrings(node any, rewrite func(string) (string, bool)) bool {
	changed := false
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if leaf, ok := child.(string); ok {
				rewritten, leafChanged := rewrite(leaf)
				if leafChanged {
					value[key] = rewritten
					changed = true
				}
				continue
			}
			if walkInputStrings(child, rewrite) {
				changed = true
			}
		}
	case []any:
		for i, child := range value {
			if leaf, ok := child.(string); ok {
				rewritten, leafChanged := rewrite(leaf)
				if leafChanged {
					value[i] = rewritten
					changed = true
				}
				continue
			}
			if walkInputStrings(child, rewrite) {
				changed = true
			}
		}
	}
	return changed
}
