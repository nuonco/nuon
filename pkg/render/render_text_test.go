package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vendorRootDomain is the expression a vendor uses to derive a nested stack
// parameter from an optional install input, falling back to a generated domain.
const vendorRootDomain = `{{ if .nuon.install.inputs.root_domain }}{{ .nuon.install.inputs.root_domain }}{{ else }}sandbox-{{ .nuon.install.id | substr 0 15 }}.installs.vendor.app{{ end }}`

func installData(inputs map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"install": map[string]interface{}{
			"id":     "1a2b3c4d5e6f7g8h9i0j",
			"inputs": inputs,
		},
	}
}

func TestRenderTextV2_ConditionalsAndPipelines(t *testing.T) {
	t.Run("input set", func(t *testing.T) {
		got, err := RenderTextV2(vendorRootDomain, installData(map[string]interface{}{
			"root_domain": "lvbl.team.example.com",
		}))
		require.NoError(t, err)
		assert.Equal(t, "lvbl.team.example.com", got)
	})

	t.Run("input empty falls back", func(t *testing.T) {
		got, err := RenderTextV2(vendorRootDomain, installData(map[string]interface{}{
			"root_domain": "",
		}))
		require.NoError(t, err)
		assert.Equal(t, "sandbox-1a2b3c4d5e6f7g8.installs.vendor.app", got)
	})

	t.Run("substr is bounded by the string length", func(t *testing.T) {
		data := map[string]interface{}{
			"install": map[string]interface{}{
				"id":     "short",
				"inputs": map[string]interface{}{"root_domain": ""},
			},
		}
		got, err := RenderTextV2(vendorRootDomain, data)
		require.NoError(t, err)
		assert.Equal(t, "sandbox-short.installs.vendor.app", got)
	})
}

// The reason this renderer exists: RenderV2 uses html/template and escapes values
// that are bound for infrastructure APIs, not browsers.
func TestRenderTextV2_DoesNotEscape(t *testing.T) {
	data := installData(map[string]interface{}{
		"connection": `user=a&pass="b"<c>`,
	})

	got, err := RenderTextV2("{{ .nuon.install.inputs.connection }}", data)
	require.NoError(t, err)
	assert.Equal(t, `user=a&pass="b"<c>`, got)

	escaped, err := RenderV2("{{ .nuon.install.inputs.connection }}", data)
	require.NoError(t, err)
	assert.NotEqual(t, got, escaped, "RenderV2 is expected to escape; if it stopped, this twin can go away")
}

func TestRenderTextV2_MissingKeyErrors(t *testing.T) {
	_, err := RenderTextV2("{{ .nuon.install.inputs.nope }}", installData(map[string]interface{}{
		"root_domain": "example.com",
	}))
	require.Error(t, err)
}

func TestRenderTextV2_PassesThroughWithoutNuonReference(t *testing.T) {
	// Matches RenderV2's short-circuit: no ".nuon", no rendering.
	got, err := RenderTextV2("{{ if true }}untouched{{ end }}", installData(nil))
	require.NoError(t, err)
	assert.Equal(t, "{{ if true }}untouched{{ end }}", got)
}

func TestValidateTextTemplate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"conditional with pipeline", vendorRootDomain, false},
		{"literal", "production", false},
		{"unterminated if", "{{ if .nuon.install.inputs.foo }}no-end", true},
		{"unknown function", "{{ .nuon.install.inputs.foo | nope }}", true},
		{"sprig function", "{{ .nuon.install.id | trunc 8 }}", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTextTemplate(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderTextStringMap(t *testing.T) {
	params := map[string]string{
		"RootDomain":  vendorRootDomain,
		"BucketName":  "vendor-{{ .nuon.install.id }}-service",
		"Environment": "production",
	}

	require.NoError(t, RenderTextStringMap(params, installData(map[string]interface{}{
		"root_domain": "",
	})))

	assert.Equal(t, map[string]string{
		"RootDomain":  "sandbox-1a2b3c4d5e6f7g8.installs.vendor.app",
		"BucketName":  "vendor-1a2b3c4d5e6f7g8h9i0j-service",
		"Environment": "production",
	}, params)
}

func TestRenderTextStringMap_ErrorNamesTheKey(t *testing.T) {
	params := map[string]string{
		"RootDomain": "{{ .nuon.install.inputs.nope }}",
	}

	err := RenderTextStringMap(params, installData(map[string]interface{}{}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RootDomain")
}
