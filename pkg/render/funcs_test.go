package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const controlPlaneSecret = "super-secret-control-plane-value"

var renderers = map[string]func(string, map[string]interface{}) (string, error){
	"RenderV2":     RenderV2,
	"RenderTextV2": RenderTextV2,
}

func TestFuncMap_EnvAccessDisabled(t *testing.T) {
	t.Setenv("NUON_RENDER_TEST_SECRET", controlPlaneSecret)

	tmpls := map[string]string{
		"env":       `{{ .nuon.install.id }}{{ env "NUON_RENDER_TEST_SECRET" }}`,
		"expandenv": `{{ .nuon.install.id }}{{ expandenv "$NUON_RENDER_TEST_SECRET" }}`,
	}

	for rname, render := range renderers {
		for fname, tmpl := range tmpls {
			t.Run(rname+"/"+fname, func(t *testing.T) {
				got, err := render(tmpl, installData(nil))
				require.Error(t, err)
				assert.Contains(t, err.Error(), "is not allowed")
				assert.NotContains(t, got, controlPlaneSecret)
			})
		}
	}

	for name, fm := range map[string]map[string]any{"TxtFuncMap": TxtFuncMap(), "HTMLFuncMap": HTMLFuncMap()} {
		t.Run(name, func(t *testing.T) {
			for _, fname := range []string{"env", "expandenv"} {
				fn, ok := fm[fname].(func(...any) (string, error))
				require.True(t, ok, "%s should be overridden", fname)

				got, err := fn("NUON_RENDER_TEST_SECRET")
				require.Error(t, err)
				assert.Empty(t, got)
			}
		})
	}
}

func TestFuncMap_DateHelpersKept(t *testing.T) {
	tmpl := `{{ .nuon.install.id }}-{{ now | date "2006" }}-{{ now | htmlDate }}`

	for rname, render := range renderers {
		t.Run(rname, func(t *testing.T) {
			got, err := render(tmpl, installData(nil))
			require.NoError(t, err)
			assert.Regexp(t, `^1a2b3c4d5e6f7g8h9i0j-\d{4}-\d{4}-\d{2}-\d{2}$`, got)
		})
	}
}
