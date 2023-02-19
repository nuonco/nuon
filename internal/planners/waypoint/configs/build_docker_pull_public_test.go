package configs

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hcl/v2/hclsimple"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_NewPublicDockerPullBuild(t *testing.T) {
	v := validator.New()
	waypointRef := generics.GetFakeObj[*planv1.WaypointRef]()
	component := generics.GetFakeObj[*componentv1.Component]()
	ecrRef := generics.GetFakeObj[*planv1.ECRRepositoryRef]()
	publicImg := generics.GetFakeObj[*PublicImageSource]()

	tests := map[string]struct {
		optsFn      func() []Option
		errExpected error
		assertFn    func(*testing.T, *publicDockerPullBuild)
	}{
		"happy path": {
			optsFn: func() []Option {
				return []Option{
					WithComponent(component),
					WithEcrRef(ecrRef),
					WithPublicImageSource(publicImg),
					WithWaypointRef(waypointRef),
				}
			},
			assertFn: func(t *testing.T, b *publicDockerPullBuild) {
				assert.Equal(t, publicImg, b.PublicImageSource)
				assert.Equal(t, component, b.Component)
				assert.Equal(t, ecrRef, b.EcrRef)
				assert.Equal(t, waypointRef, b.WaypointRef)
			},
		},
		"missing public image source": {
			optsFn: func() []Option {
				return []Option{
					WithComponent(component),
					WithEcrRef(ecrRef),
					WithWaypointRef(waypointRef),
				}
			},
			errExpected: fmt.Errorf("public image source not provided"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := test.optsFn()
			bld, err := NewPublicDockerPullBuild(v, opts...)
			if test.errExpected != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, bld)
		})
	}
}

func Test_PublicDockerPullBuild(t *testing.T) {
	v := validator.New()
	waypointRef := generics.GetFakeObj[*planv1.WaypointRef]()
	component := generics.GetFakeObj[*componentv1.Component]()
	ecrRef := generics.GetFakeObj[*planv1.ECRRepositoryRef]()
	publicImg := generics.GetFakeObj[*PublicImageSource]()

	tests := map[string]struct {
		builderFn func(*testing.T) *publicDockerPullBuild
		assertFn  func(*testing.T, waypointv1.Hcl_Format, []byte)
	}{
		"happy path": {
			builderFn: func(t *testing.T) *publicDockerPullBuild {
				bld, err := NewPublicDockerPullBuild(v,
					WithComponent(component),
					WithEcrRef(ecrRef),
					WithPublicImageSource(publicImg),
					WithWaypointRef(waypointRef))
				assert.NoError(t, err)
				return bld
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
				assert.Equal(t, "docker-pull", appBlock.Build[0].Use[0].Name)
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
