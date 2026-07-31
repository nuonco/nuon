package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	release "helm.sh/helm/v4/pkg/release/v1"
)

// A drift check renders the chart with PlanOnly, so the kinds are knowable
// without applying — that is what lets a component pick up its custom resource
// kinds without being redeployed.
func TestRenderedManifestUsesPlanOutput(t *testing.T) {
	const applied = "kind: NodePool\n"
	const planned = "kind: EC2NodeClass\n"

	t.Run("an applied release wins", func(t *testing.T) {
		got := renderedManifest(&release.Release{Manifest: applied},
			HelmPlanContents{Op: "upgrade", TemplateOutput: planned})
		assert.Equal(t, applied, got)
	})

	t.Run("plan-only falls back to the template output", func(t *testing.T) {
		got := renderedManifest(nil, HelmPlanContents{Op: "upgrade", TemplateOutput: planned})
		assert.Equal(t, planned, got)
	})

	t.Run("an uninstall plan records nothing", func(t *testing.T) {
		got := renderedManifest(nil, HelmPlanContents{Op: "uninstall", TemplateOutput: planned})
		assert.Empty(t, got, "objects being removed are not owned kinds")
	})

	t.Run("no release and no output", func(t *testing.T) {
		assert.Empty(t, renderedManifest(nil, HelmPlanContents{Op: "install"}))
	})
}
