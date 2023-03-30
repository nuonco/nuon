package deployv1

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestHelmRepoConfig(t *testing.T) {
	t.Run("source config", func(t *testing.T) {
		obj := HelmRepoConfig{
			ChartRepo:    "https://artifacthub.io/org/helm-repo",
			ChartName:    "kewl-chart-name",
			ChartVersion: "v0.0.0",
		}
		assert.NoError(t, obj.Validate())
	})

	t.Run("invalid source config", func(t *testing.T) {
		obj := HelmRepoConfig{
			ChartRepo:    "",
			ChartName:    "kewl-chart-name",
			ChartVersion: "v0.0.0",
		}
		assert.Error(t, obj.Validate())
	})

	t.Run("no source config", func(t *testing.T) {
		obj := HelmRepoConfig{
			ChartRepo:    "https://artifacthub.io/org/helm-repo",
			ChartName:    "kewl-chart-name",
			ChartVersion: "v0.0.0",

			ImageRepoValuesKey: types.ToPtr("image.repo"),
			ImageTagValuesKey:  types.ToPtr("image.tag"),
		}
		assert.NoError(t, obj.Validate())
	})
}
