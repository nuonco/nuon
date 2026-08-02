package build

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type KubernetesContextInput struct {
	Name          string
	ComponentName string
}

func KubernetesContextInputsFromConfig(contexts *config.KubernetesContextsConfig) []KubernetesContextInput {
	if contexts == nil {
		return nil
	}
	out := make([]KubernetesContextInput, 0, len(contexts.Contexts))
	for _, c := range contexts.Contexts {
		out = append(out, KubernetesContextInput{Name: c.Name, ComponentName: c.Component})
	}
	return out
}

// KubernetesContextsConfig resolves each context's source component name against
// componentIDByName, which the caller looks up. The name is persisted alongside
// the ID so the binding stays intelligible if the component is renamed.
func KubernetesContextsConfig(contexts []KubernetesContextInput, componentIDByName map[string]string, appID, appConfigID string) (*app.AppKubernetesContextsConfig, error) {
	children := make([]app.AppKubernetesContextConfig, 0, len(contexts))
	for _, c := range contexts {
		compID, ok := componentIDByName[c.ComponentName]
		if !ok {
			return nil, fmt.Errorf("kubernetes_context %q references unknown component %q", c.Name, c.ComponentName)
		}
		children = append(children, app.AppKubernetesContextConfig{
			AppID:               appID,
			AppConfigID:         appConfigID,
			Name:                c.Name,
			SourceComponentName: c.ComponentName,
			SourceComponentID:   compID,
		})
	}

	return &app.AppKubernetesContextsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Contexts:    children,
	}, nil
}
