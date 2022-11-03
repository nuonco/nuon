package waypoint

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentGenerateHCL(t *testing.T) {
	tests := map[string]struct {
		component   Component
		expected    []byte
		errExpected error
		version     int
	}{
		"happy path build": {
			component: Component{
				Name:              "component-name",
				ContainerImageURL: "container-url",
			},
			expected: []byte(`app "component-name" {
  build {
    use "docker-pull" {
			registry {
				image: "container-url",
				tag: "latest",
			}
    }
  }
}`),
			version: HCLBuild,
		},
		"happy path deploy": {
			component: Component{
				Name:              "component-name",
				ContainerImageURL: "container-url",
			},
			expected: []byte(`app "component-name" {
  deploy {
    use "kubernetes" {
    }
  }
}`),
			version: HCLDeploy,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := test.component.GenerateHCL(test.version)
			buffer := new(bytes.Buffer)
			// use json.Compact because newlines and tabs are a pain
			assert.Equal(t, json.Compact(buffer, test.expected), json.Compact(buffer, actual.Bytes()))
		})
	}
}
