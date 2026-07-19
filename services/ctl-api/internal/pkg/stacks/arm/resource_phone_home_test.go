package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestGetPhoneHomeResource_CustomStackOutputs(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()

	res := tmpl.getPhoneHomeResource(inp, []customDeploymentOutputs{
		{StackName: "preview_bucket", DeploymentName: "PreviewBucket", OutputKeys: []string{"bucketName", "bucketUrl"}},
	})

	props := res["properties"].(map[string]any)
	script := props["scriptContent"].(string)
	for _, want := range []string{
		`"custom_nested_stacks"`,
		`"preview_bucket"`,
		`"bucketName": "$CUSTOM_PREVIEW_BUCKET_BUCKETNAME"`,
		`"bucketUrl": "$CUSTOM_PREVIEW_BUCKET_BUCKETURL"`,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %s", want)
		}
	}

	found := false
	for _, env := range props["environmentVariables"].([]map[string]any) {
		if env["name"] == "CUSTOM_PREVIEW_BUCKET_BUCKETNAME" {
			found = true
			if env["value"] != "[string(reference('PreviewBucket').outputs.bucketName.value)]" {
				t.Errorf("unexpected env value: %v", env["value"])
			}
		}
	}
	if !found {
		t.Error("CUSTOM_PREVIEW_BUCKET_BUCKETNAME env var not found")
	}

	deps := res["dependsOn"].([]string)
	if !strings.Contains(strings.Join(deps, ","), "PreviewBucket") {
		t.Errorf("dependsOn missing PreviewBucket: %v", deps)
	}

	resBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	assertNoNestedBrackets(t, resBytes)
}
