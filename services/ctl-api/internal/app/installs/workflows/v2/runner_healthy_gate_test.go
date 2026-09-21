package v2

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEveryRunnerHealthGateHasExpectedMode(t *testing.T) {
	want := map[string][]string{
		"app_branch_config_update.go": {"ModeRequireActive", "ModeStartup"},
		"deprovision.go":              {"ModeRequireActive"},
		"deprovision_sandbox.go":      {"ModeRequireActive"},
		"deploy_components.go":        {"ModeRequireActive"},
		"manual_deploy.go":            {"ModeRequireActive"},
		"provision.go":                {"ModeStartup"},
		"recover_helm_release.go":     {"ModeRequireActive"},
		"reprovision_sandbox.go":      {"ModeRequireActive"},
		"reprovision_stack.go":        {"ModeStartup"},
		"shared_helpers.go":           {"ModeRequireActive"},
		"teardown_component.go":       {"ModeRequireActive"},
		"teardown_components.go":      {"ModeRequireActive"},
		"update_input.go":             {"ModeRequireActive"},
	}

	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	got := make(map[string][]string)
	for _, filename := range files {
		if filepath.Ext(filename) != ".go" || strings.HasSuffix(filename, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
		require.NoError(t, err)
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Signal" {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "awaitrunnerhealthy" {
				return true
			}
			hasMode := false
			for _, elt := range lit.Elts {
				field, ok := elt.(*ast.KeyValueExpr)
				if !ok || field.Key.(*ast.Ident).Name != "Mode" {
					continue
				}
				hasMode = true
				mode, ok := field.Value.(*ast.SelectorExpr)
				require.True(t, ok, "%s readiness mode must use an exported constant", filename)
				got[filepath.Base(filename)] = append(got[filepath.Base(filename)], mode.Sel.Name)
			}
			require.True(t, hasMode, "%s runner health gate must set an explicit mode", filename)
			return true
		})
	}

	require.Equal(t, want, got)
}

func TestComposedWorkflowsDoNotDuplicateRunnerHealthGate(t *testing.T) {
	tests := []struct {
		filename string
		callee   string
		want     string
	}{
		{filename: "deprovision.go", callee: "teardownComponents", want: "false"},
		{filename: "teardown_components.go", callee: "teardownComponents", want: "true"},
		{filename: "update_input.go", callee: "getSandboxReprovisionSteps", want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), tt.filename, nil, 0)
			require.NoError(t, err)

			var got []string
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				fn, ok := call.Fun.(*ast.Ident)
				if !ok || fn.Name != tt.callee || len(call.Args) == 0 {
					return true
				}
				arg, ok := call.Args[len(call.Args)-1].(*ast.Ident)
				require.True(t, ok, "%s must pass an explicit gate decision to %s", tt.filename, tt.callee)
				got = append(got, arg.Name)
				return true
			})

			require.Equal(t, []string{tt.want}, got)
		})
	}
}
