package validate

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
	"gopkg.in/yaml.v2"
)

func ValidatePolicies(a *config.AppConfig) error {
	if a.Policies == nil || len(a.Policies.Policies) < 1 {
		return nil
	}

	for idx, policy := range a.Policies.Policies {
		var obj map[string]any

		if err := yaml.Unmarshal([]byte(policy.Contents), &obj); err != nil {
			return config.ErrConfig{
				Description: fmt.Sprintf("policy %d (%s) was invalid", idx, policy.Type),
				Err:         err,
			}
		}
	}

	return nil
}
