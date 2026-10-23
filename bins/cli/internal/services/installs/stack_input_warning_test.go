package installs

import (
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestInputNamesReferencedByStack(t *testing.T) {
	appCfg := &models.AppAppConfig{
		Stack: &models.AppAppStackConfig{
			Name: "{{ .nuon.install.inputs.stack_name }}",
			CustomNestedStacks: []*models.ConfigCustomNestedStack{
				{
					Name: "database",
					Parameters: map[string]string{
						"InstanceClass": "{{.nuon.install.inputs.instance_class}}",
						"Static":        "db.t3.micro",
					},
				},
			},
		},
		Permissions: &models.AppAppPermissionsConfig{
			AwsIamRoles: []*models.AppAppAWSIAMRoleConfig{
				{
					Policies: []*models.AppAppAWSIAMPolicyConfig{
						{Contents: `{"Sid":"{{ .nuon.inputs.inputs.policy_sid }}"}`},
					},
				},
			},
		},
		Sandbox: &models.AppAppSandboxConfig{
			AwsRegionType: "{{ .nuon.install.inputs.sandbox_only }}",
		},
		Secrets: &models.AppAppSecretsConfig{},
	}

	installCfg := &models.AppInstallConfig{
		RunnerNestedTemplateURL: "{{ .nuon.install.inputs.runner_template }}",
	}

	got := inputNamesReferencedByStack(appCfg, installCfg)
	want := []string{"instance_class", "policy_sid", "runner_template", "stack_name"}
	if len(got) != len(want) {
		t.Fatalf("names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("names = %v, want %v", got, want)
		}
	}
}

func TestInputNamesReferencedByStackIgnoresSandbox(t *testing.T) {
	appCfg := &models.AppAppConfig{
		Sandbox: &models.AppAppSandboxConfig{
			AwsRegionType: "{{ .nuon.install.inputs.sandbox_only }}",
		},
	}
	if got := inputNamesReferencedByStack(appCfg, nil); len(got) != 0 {
		t.Fatalf("names = %v, want none", got)
	}
}

func TestFormatStackReprovisionWarning(t *testing.T) {
	if got := formatStackReprovisionWarning(nil); got != "" {
		t.Fatalf("empty = %q", got)
	}
	got := formatStackReprovisionWarning([]string{"instance_class"})
	want := "The install stack uses instance_class. Run nuon installs stacks reprovision to apply this change to the stack"
	if got != want {
		t.Fatalf("singular = %q, want %q", got, want)
	}
	got = formatStackReprovisionWarning([]string{"instance_class", "stack_name"})
	want = "The install stack uses instance_class and stack_name. Run nuon installs stacks reprovision to apply these changes to the stack"
	if got != want {
		t.Fatalf("plural = %q, want %q", got, want)
	}
}

func TestChangedStackInputs(t *testing.T) {
	changed := map[string]bool{"instance_class": true, "unchanged": false}
	got := changedStackInputs(changed, []string{"instance_class", "stack_name", "unchanged"})
	if len(got) != 1 || got[0] != "instance_class" {
		t.Fatalf("changed = %v", got)
	}
}
