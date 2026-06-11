package helm

import (
	"testing"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

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
