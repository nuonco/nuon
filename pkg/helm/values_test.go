package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/action"
	rcommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func TestUpgradeValues(t *testing.T) {
	for _, status := range []rcommon.Status{rcommon.StatusDeployed, rcommon.StatusFailed, rcommon.StatusSuperseded} {
		t.Run(string(status), func(t *testing.T) {
			store := storage.Init(driver.NewMemory())
			current := &release.Release{Name: "acme", Namespace: "default", Version: 1, Info: &release.Info{Status: status}, Config: map[string]interface{}{"enabled": true}}
			require.NoError(t, store.Create(current))
			cfg := &action.Configuration{Releases: store}
			values, err := UpgradeValues(cfg, current, nil)
			require.NoError(t, err)
			require.Equal(t, map[string]interface{}{"enabled": true}, values)
			values, err = UpgradeValues(cfg, current, map[string]interface{}{"enabled": false})
			require.NoError(t, err)
			require.Equal(t, map[string]interface{}{"enabled": false}, values)
			if status == rcommon.StatusDeployed {
				last := &release.Release{Name: "acme", Namespace: "default", Version: 2, Info: &release.Info{Status: rcommon.StatusFailed}, Config: map[string]interface{}{"enabled": false}}
				require.NoError(t, store.Create(last))
				values, err = UpgradeValues(cfg, last, nil)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"enabled": true}, values)
			}
		})
	}
}

func TestChartValues_OverrideWinsOverFilesAndSet(t *testing.T) {
	valuesFile := `
replicaCount: 1
image:
  repository: app
  tag: v1
resources:
  requests:
    cpu: "1"
`
	set := []plantypes.HelmValue{
		{Name: "image.tag", Value: "v2"}, // --set normally wins over files
		{Name: "replicaCount", Value: "3"},
	}
	override := `
replicaCount: 5
image:
  tag: v3-override
`

	out, err := ChartValues([]string{valuesFile}, set, override)
	if err != nil {
		t.Fatalf("ChartValues error: %v", err)
	}

	// override beats both the values file and the --set value
	if got := out["replicaCount"]; got != float64(5) && got != int64(5) && got != 5 {
		t.Errorf("replicaCount = %v (%T), want 5", got, got)
	}

	image, _ := out["image"].(map[string]interface{})
	if image == nil {
		t.Fatalf("image map missing: %#v", out["image"])
	}
	if image["tag"] != "v3-override" {
		t.Errorf("image.tag = %v, want v3-override (override should beat --set)", image["tag"])
	}
	// untouched keys from the values file survive the deep merge
	if image["repository"] != "app" {
		t.Errorf("image.repository = %v, want app (sparse merge should preserve)", image["repository"])
	}
}

func TestChartValues_EmptyOverrideIsNoop(t *testing.T) {
	valuesFile := "replicaCount: 1\nimage:\n  tag: v1\n"
	set := []plantypes.HelmValue{{Name: "image.tag", Value: "v2"}}

	base, err := ChartValues([]string{valuesFile}, set, "")
	if err != nil {
		t.Fatalf("base error: %v", err)
	}
	withBlank, err := ChartValues([]string{valuesFile}, set, "   \n")
	if err != nil {
		t.Fatalf("blank override error: %v", err)
	}

	image, _ := base["image"].(map[string]interface{})
	if image["tag"] != "v2" {
		t.Errorf("base image.tag = %v, want v2 (--set wins with no override)", image["tag"])
	}
	imageBlank, _ := withBlank["image"].(map[string]interface{})
	if imageBlank["tag"] != "v2" {
		t.Errorf("blank-override image.tag = %v, want v2 (whitespace override is no-op)", imageBlank["tag"])
	}
}
