package customermanaged

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubstituteInputValues(t *testing.T) {
	plan := map[string]any{
		"deploy_plan": map[string]any{
			"helm_values": map[string]any{
				"domain":   InputPlaceholder("app_domain"),
				"combined": "https://" + InputPlaceholder("app_domain") + "/api",
				"other":    "untouched",
			},
			"list": []any{InputPlaceholder("replicas"), "static"},
		},
	}
	SubstituteInputValues(plan, map[string]string{
		"app_domain": "app.example.com",
		"replicas":   "3",
	})

	deploy := plan["deploy_plan"].(map[string]any)
	values := deploy["helm_values"].(map[string]any)
	require.Equal(t, "app.example.com", values["domain"])
	require.Equal(t, "https://app.example.com/api", values["combined"])
	require.Equal(t, "untouched", values["other"])
	require.Equal(t, "3", deploy["list"].([]any)[0])
}

func TestUnresolvedInputPlaceholders(t *testing.T) {
	specs := []InputSpec{
		{Name: "app_domain", Bindable: true},
		{Name: "replicas", Bindable: true},
		{Name: "unused", Bindable: true},
	}
	plan := map[string]any{
		"a": InputPlaceholder("app_domain"),
		"b": []any{map[string]any{"c": "prefix-" + InputPlaceholder("replicas")}},
	}
	require.Equal(t, []string{"app_domain", "replicas"}, UnresolvedInputPlaceholders(plan, specs))

	SubstituteInputValues(plan, map[string]string{"app_domain": "x", "replicas": "1"})
	require.Empty(t, UnresolvedInputPlaceholders(plan, specs))
}

func TestValidateInputValues(t *testing.T) {
	specs := []InputSpec{
		{Name: "app_domain", Type: "string", Required: true, Bindable: true},
		{Name: "replicas", Type: "number", Bindable: true},
		{Name: "enabled", Type: "bool", Bindable: true},
		{Name: "api_key", Type: "string", Secret: true},
		{Name: "baked", Type: "string", Bindable: false},
		{Name: "with_default", Type: "string", Required: true, Bindable: true, Default: "fallback"},
	}

	require.NoError(t, ValidateInputValues(specs, map[string]string{
		"app_domain": "app.example.com",
		"replicas":   "3",
		"enabled":    "true",
	}))

	err := ValidateInputValues(specs, map[string]string{"nope": "x"})
	require.ErrorContains(t, err, `input "nope" is not declared`)
	require.ErrorContains(t, err, `required input "app_domain" has no value`)

	err = ValidateInputValues(specs, map[string]string{"app_domain": "x", "api_key": "s3cret"})
	require.ErrorContains(t, err, "secrets are not supported")

	err = ValidateInputValues(specs, map[string]string{"app_domain": "x", "baked": "y"})
	require.ErrorContains(t, err, "not late-bindable")

	err = ValidateInputValues(specs, map[string]string{"app_domain": "x", "replicas": "many"})
	require.ErrorContains(t, err, `input "replicas" must be a number`)

	err = ValidateInputValues(specs, map[string]string{"app_domain": "x", "enabled": "maybe"})
	require.ErrorContains(t, err, `input "enabled" must be a boolean`)
}

func TestResolveInputValues(t *testing.T) {
	specs := []InputSpec{
		{Name: "app_domain", Bindable: true},
		{Name: "with_default", Bindable: true, Default: "fallback"},
		{Name: "overridden_default", Bindable: true, Default: "old"},
		{Name: "secret", Bindable: true, Secret: true, Default: "nope"},
		{Name: "baked", Bindable: false, Default: "nope"},
		{Name: "unset_optional", Bindable: true},
		{Name: "unset_required", Bindable: true, Required: true},
	}
	resolved := ResolveInputValues(specs, map[string]string{
		"app_domain":         "app.example.com",
		"overridden_default": "new",
		"secret":             "leak",
		"baked":              "ignored",
	})
	require.Equal(t, map[string]string{
		"app_domain":         "app.example.com",
		"with_default":       "fallback",
		"overridden_default": "new",
		"unset_optional":     "",
	}, resolved)
}

func TestInputPlaceholderShape(t *testing.T) {
	token := InputPlaceholder("my_input_2")
	require.True(t, strings.HasPrefix(token, "__NUON_INPUT_"))
	require.True(t, strings.HasSuffix(token, "__"))
}
