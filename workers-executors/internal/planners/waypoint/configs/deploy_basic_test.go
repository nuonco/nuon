package configs

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hcl/v2/hclsimple"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_basicDeployBuilder_Render(t *testing.T) {
	v := validator.New()
	waypointRef := generics.GetFakeObj[*planv1.WaypointRef]()
	component := generics.GetFakeObj[*componentv1.Component]()
	ecrRef := generics.GetFakeObj[*planv1.ECRRepositoryRef]()

	tests := map[string]struct {
		builderFn func(*testing.T) *basicDeploy
		assertFn  func(*testing.T, waypointv1.Hcl_Format, []byte)
	}{
		"happy path": {
			builderFn: func(t *testing.T) *basicDeploy {
				dep, err := NewBasicDeploy(v,
					WithComponent(component),
					WithEcrRef(ecrRef),
					WithWaypointRef(waypointRef))
				assert.NoError(t, err)
				return dep
			},
			assertFn: func(t *testing.T, fmt waypointv1.Hcl_Format, byts []byte) {
				assert.Equal(t, fmt, waypointv1.Hcl_HCL)
				assert.NotNil(t, byts)

				var wpConfig waypointConfig
				err := hclsimple.Decode(defaultWaypointConfigFilename, byts, nil, &wpConfig)
				assert.NoError(t, err)

				assert.Equal(t, waypointRef.Project, wpConfig.Project)
				assert.Equal(t, waypointRef.App, wpConfig.App[0].Name)

				appBlock := wpConfig.App[0]
				assert.Equal(t, "kubernetes", appBlock.Deploy[0].Use[0].Name)
				assert.Equal(t, "aws-ecr", appBlock.Build[0].Registry[0].Use.Name)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bldr := test.builderFn(t)
			byts, fmt, err := bldr.Render()
			assert.NoError(t, err)
			test.assertFn(t, fmt, byts)
		})
	}
}
