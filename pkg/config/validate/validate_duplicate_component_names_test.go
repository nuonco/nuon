package validate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidateDuplicateComponentNames_IdentityProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		count := hegel.Draw(t, hegel.Integers(1, 8))
		suffix := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		cfg := &config.AppConfig{}
		for i := range count {
			name := fmt.Sprintf("component_%d_%s", i, suffix)
			cfg.Components = append(cfg.Components, &config.Component{Name: name, Type: config.JobComponentType, Job: &config.JobComponentConfig{}})
			cfg.Actions = append(cfg.Actions, &config.ActionConfig{Name: name})
		}
		require.NoError(t, ValidateDuplicateComponentNames(cfg), "same names in different resource namespaces are valid")
		index := hegel.Draw(t, hegel.Integers(0, count-1))
		duplicate := *cfg.Components[index]
		for _, samePayload := range []bool{false, true} {
			if !samePayload {
				duplicate.Type, duplicate.Job = config.TerraformModuleComponentType, nil
				duplicate.TerraformModule = &config.TerraformModuleComponentConfig{TerraformVersion: "1.6.0"}
			} else {
				duplicate = *cfg.Components[index]
			}
			position := hegel.Draw(t, hegel.Integers(0, count))
			withDuplicate := *cfg
			withDuplicate.Components = append(config.ComponentList{}, cfg.Components[:position]...)
			withDuplicate.Components = append(withDuplicate.Components, &duplicate)
			withDuplicate.Components = append(withDuplicate.Components, cfg.Components[position:]...)
			require.ErrorContains(t, ValidateDuplicateComponentNames(&withDuplicate), duplicate.Name)
		}
		require.Len(t, cfg.Components, count)
		require.NoError(t, ValidateDuplicateComponentNames(cfg))
	})
}
