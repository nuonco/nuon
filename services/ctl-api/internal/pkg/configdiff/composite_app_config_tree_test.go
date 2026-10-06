package configdiff

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
)

func TestSpliceCompositeAppConfigTreeUsesEntityBaselines(t *testing.T) {
	newCfg := &config.AppConfig{
		Description: "from-new",
		Sandbox:     &config.AppSandboxConfig{TerraformVersion: "1.6.0"},
		Stack:       &config.StackConfig{Name: "primary", Type: "aws-cloudformation"},
		Runner:      &config.AppRunnerConfig{RunnerType: "ecs"},
		Components: config.ComponentList{
			&config.Component{Name: "api", Type: config.JobComponentType, VarName: "new"},
			&config.Component{Name: "web", Type: config.JobComponentType, VarName: "web"},
		},
	}
	tree := spliceCompositeAppConfigTree(compositeTreeInputs{
		New: newCfg,
		Intermediates: map[string]*config.AppConfig{
			"cfg-sandbox": {
				Description: "from-sandbox",
				Sandbox:     &config.AppSandboxConfig{TerraformVersion: "1.5.0"},
			},
			"cfg-stack": {
				Stack:  &config.StackConfig{Name: "old-stack", Type: "aws-cloudformation"},
				Runner: &config.AppRunnerConfig{RunnerType: "eks"},
			},
			"cfg-api": {
				Components: config.ComponentList{
					&config.Component{Name: "api", Type: config.JobComponentType, VarName: "old"},
					&config.Component{Name: "legacy", Type: config.JobComponentType, VarName: "gone"},
				},
			},
		},
		Baselines: CompositeBaselines{
			Components: map[string]string{"cmp-api": "cfg-api"},
			Sandbox:    "cfg-sandbox",
			Stack:      "cfg-stack",
		},
		ComponentIDsByName: map[string]string{
			"api":    "cmp-api",
			"legacy": "cmp-legacy",
		},
	})

	if op := fieldOp(t, tree, "description"); op != pkgdiff.OpAdd {
		t.Fatalf("description op = %s, want add against an empty install baseline", op)
	}
	if op := fieldOp(t, topChild(tree, "sandbox"), "terraform_version"); op != pkgdiff.OpChange {
		t.Fatalf("sandbox terraform_version op = %s, want change against the sandbox baseline", op)
	}
	if op := fieldOp(t, topChild(tree, "stack"), "name"); op != pkgdiff.OpChange {
		t.Fatalf("stack name op = %s, want change against the stack baseline", op)
	}
	if op := fieldOp(t, topChild(tree, "runner"), "runner_type"); op != pkgdiff.OpChange {
		t.Fatalf("runner type op = %s, want change against the stack baseline", op)
	}
	if op := fieldOp(t, componentNode(tree, "api"), "var_name"); op != pkgdiff.OpChange {
		t.Fatalf("api var_name op = %s, want change against the component baseline", op)
	}
	if op := fieldOp(t, componentNode(tree, "web"), "var_name"); op != pkgdiff.OpAdd {
		t.Fatalf("web var_name op = %s, want add because web has no applied config", op)
	}
	if componentNode(tree, "legacy") == nil {
		t.Fatal("legacy component missing; it exists only on another component baseline and should show as removed")
	}
	if op := fieldOp(t, componentNode(tree, "legacy"), "var_name"); op != pkgdiff.OpRemove {
		t.Fatalf("legacy var_name op = %s, want remove", op)
	}
}

func fieldOp(t *testing.T, node *pkgdiff.Diff, key string) pkgdiff.Op {
	t.Helper()
	child := topChild(node, key)
	if child == nil || child.Diff == nil {
		t.Fatalf("missing field %s", key)
	}
	return child.Diff.Op
}
