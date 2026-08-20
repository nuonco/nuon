package arm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// The golden files were generated from the commit before the deployment-scope
// work began, so this asserts the real backwards-compatibility contract: an app
// that has not opted in renders exactly the bytes it rendered before.
//
// A checksum alone would only say that drift happened. Committing the bytes means
// the diff shows what drifted, which matters because the subscription-scope work
// changes how these resources are assembled — resource ordering included.
//
// Regenerate deliberately, never reflexively: UPDATE_GOLDEN=1 go test ./...
// and read the diff. A change here re-renders every existing Azure install's
// template on its next reprovision.
func TestGetAzureTemplate_RGScopeMatchesGolden(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	for _, tc := range []struct {
		golden string
		build  func() *stacks.TemplateInput
	}{
		{"rg_scope_minimal.json", minimalTemplateInput},
		{"rg_scope_roles.json", azureRolesTemplateInput},
	} {
		t.Run(tc.golden, func(t *testing.T) {
			got, _, err := tmpl.Template(tc.build())
			if err != nil {
				t.Fatalf("render: %v", err)
			}

			path := filepath.Join("testdata", tc.golden)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				if err := os.WriteFile(path, got, 0o644); err != nil {
					t.Fatalf("update golden: %v", err)
				}
				t.Logf("updated %s", path)
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("rendered template drifted from %s; run UPDATE_GOLDEN=1 and review the diff", path)
			}
		})
	}
}
