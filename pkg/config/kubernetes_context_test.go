package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveKubernetesContexts(t *testing.T) {
	tests := map[string]struct {
		context  *KubernetesContext
		wantErr  string
		wantDeps []string
	}{
		"peer component context adds a dependency": {
			context:  &KubernetesContext{Name: "data-cluster", Component: "data-eks"},
			wantDeps: []string{"data-eks"},
		},
		"in-cluster context adds no dependency": {
			context: &KubernetesContext{Name: "runner-cluster", InCluster: true},
		},
		"in-cluster context rejects a component": {
			context: &KubernetesContext{Name: "runner-cluster", Component: "data-eks", InCluster: true},
			wantErr: "component can not be set when in_cluster is true",
		},
		"context without component or in_cluster is rejected": {
			context: &KubernetesContext{Name: "data-cluster"},
			wantErr: "component is required unless in_cluster is true",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			api := &Component{Name: "acme-api", Type: HelmChartComponentType, KubernetesContext: tc.context.Name}
			cfg := &AppConfig{
				Components: ComponentList{
					{Name: "data-eks", Type: TerraformModuleComponentType},
					api,
				},
				KubernetesContexts: &KubernetesContextsConfig{Contexts: []*KubernetesContext{tc.context}},
			}

			err := cfg.resolveKubernetesContexts()
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tc.wantDeps, api.Dependencies)
		})
	}
}
