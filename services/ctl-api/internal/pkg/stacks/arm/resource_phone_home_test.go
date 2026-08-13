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

func TestGetPhoneHomeResource_RunnerIdentityPrincipalID(t *testing.T) {
	t.Run("reported when a runner deployment exists", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		res := tmpl.getPhoneHomeResource(minimalTemplateInput(), nil)

		props := res["properties"].(map[string]any)
		script := props["scriptContent"].(string)
		if !strings.Contains(script, `"runner_identity_principal_id": "$RUNNER_IDENTITY_PRINCIPAL_ID"`) {
			t.Error("payload missing runner_identity_principal_id")
		}

		var found string
		for _, ev := range props["environmentVariables"].([]map[string]any) {
			if ev["name"] == "RUNNER_IDENTITY_PRINCIPAL_ID" {
				found = ev["value"].(string)
			}
		}
		if want := "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"; found != want {
			t.Errorf("env var = %q, want %q", found, want)
		}
	})

	t.Run("omitted for local runners", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{UseLocalRunners: true}}
		res := tmpl.getPhoneHomeResource(minimalTemplateInput(), nil)

		script := res["properties"].(map[string]any)["scriptContent"].(string)
		if strings.Contains(script, "RUNNER_IDENTITY_PRINCIPAL_ID") {
			t.Error("local runners have no runnerDeployment to reference")
		}
	})
}
