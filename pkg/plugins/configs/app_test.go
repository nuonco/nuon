package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type hclRenderer interface {
	ToHCL() []byte
}

func TestToHCL(t *testing.T) {
	tests := map[string]struct {
		appFn    func() hclRenderer
		assertFn func(*testing.T, []byte)
	}{
		"basic app": {
			appFn: func() hclRenderer {
				type build struct {
					Name string `hcl:"name,label"`
				}
				type deploy struct {
					Name string `hcl:"name,label"`
				}

				return &Apps[build, deploy]{
					Project: "project-id",
					App: App[build, deploy]{
						Name: "app-id",
						Build: build{
							Name: "build-plugin",
						},
						Deploy: deploy{
							Name: "deploy-plugin",
						},
					},
				}
			},
			assertFn: func(t *testing.T, byts []byte) {
				contains := []string{"project-id", "app-id", "build-plugin", "deploy-plugin"}
				for _, contain := range contains {
					assert.Contains(t, string(byts), contain)
				}
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			app := test.appFn()
			byts := app.ToHCL()
			test.assertFn(t, byts)
		})
	}
}
