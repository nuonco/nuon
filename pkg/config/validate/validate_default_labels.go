package validate

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
)

// ValidateDefaultLabels enforces the same label rules as install labels, and
// rejects per-install label keys that collide with a default — the app-level
// block is the only place a default key may be set.
func ValidateDefaultLabels(a *config.AppConfig) error {
	for key, val := range a.DefaultLabels {
		if strings.Contains(key, "{{") {
			return config.ErrConfig{
				Description: fmt.Sprintf("default label key %q must not use the interpolation syntax", key),
				Err:         fmt.Errorf("default label key %q is templated; only label values may be templated", key),
			}
		}
		if !labels.IsTemplatedValue(val) {
			continue
		}
		if err := render.ValidateTextTemplate(val); err != nil {
			return config.ErrConfig{
				Description: fmt.Sprintf("default label %q is not a valid template", key),
				Err:         fmt.Errorf("default label %q template: %w", key, err),
			}
		}
	}

	if len(a.DefaultLabels) == 0 {
		return nil
	}

	for _, install := range a.Installs {
		for key := range install.Labels {
			if _, ok := a.DefaultLabels[key]; ok {
				return config.ErrConfig{
					Description: fmt.Sprintf("install %q label %q is a default label; set it only in default_labels", install.Name, key),
					Err:         fmt.Errorf("install %q label %q collides with a default label", install.Name, key),
				}
			}
		}
	}

	return nil
}
