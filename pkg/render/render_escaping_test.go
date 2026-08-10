package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pemPublicKey is the shape that first exposed the escaping bug: a PEM body is
// standard base64, so it contains "+" roughly as often as any other character.
const pemPublicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtC28SUqEM2D1AUBKuZyG
kSht4TZkX6EdxMjPVO6X/ElGDeKxxBkMWEEGOOARWrcaeK13K3+mzLtIEWiX5QOQ
SFTHTcRsDUoH1fuEodZmFGUCAwEAAQ==
-----END PUBLIC KEY-----
`

// helmComponentConfig mirrors the shape of app.HelmComponentConfig: the values
// files are carried as a []string of file contents tagged for templating.
type helmComponentConfig struct {
	ChartName   string            `features:"template"`
	ValuesFiles []string          `features:"template"`
	Values      map[string]string `features:"template"`
}

func escapingData() map[string]interface{} {
	return map[string]interface{}{
		"install": map[string]interface{}{
			"inputs": map[string]interface{}{
				"lovable_public_api_key": pemPublicKey,
				"bootstrap_token":        "bst_ws-abc_dGVzdC10b2tlbg",
				"connection":             `user=a&pass="b"<c>'d'`,
			},
		},
	}
}

// A helm values file carrying a PEM through an input must survive byte for byte.
// Through html/template the "+" became "&#43;" and the sandbox failed to parse
// the key.
func TestRenderStruct_HelmValuesFileKeepsPEMIntact(t *testing.T) {
	cfg := &helmComponentConfig{
		ChartName: "sandbox-cluster",
		ValuesFiles: []string{
			"sandboxAuth:\n  lovablePublicKey: |-{{ .nuon.install.inputs.lovable_public_api_key | nindent 4 }}\n",
		},
	}

	require.NoError(t, RenderStruct(cfg, escapingData()))

	assert.Contains(t, cfg.ValuesFiles[0], "eK13K3+mzLtIEWiX5QOQ")
	assert.NotContains(t, cfg.ValuesFiles[0], "&#43;")
}

func TestRenderStruct_DoesNotEscapeStringFields(t *testing.T) {
	cfg := &helmComponentConfig{
		ChartName: "{{ .nuon.install.inputs.connection }}",
	}

	require.NoError(t, RenderStruct(cfg, escapingData()))

	assert.Equal(t, `user=a&pass="b"<c>'d'`, cfg.ChartName)
}

func TestRenderMap_DoesNotEscapeValues(t *testing.T) {
	vals := map[string]string{
		"relay.bootstrapToken": "{{ .nuon.install.inputs.bootstrap_token }}",
		"proxy.connection":     "{{ .nuon.install.inputs.connection }}",
	}

	require.NoError(t, RenderMap(&vals, escapingData()))

	assert.Equal(t, "bst_ws-abc_dGVzdC10b2tlbg", vals["relay.bootstrapToken"])
	assert.Equal(t, `user=a&pass="b"<c>'d'`, vals["proxy.connection"])
}
