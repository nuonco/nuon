package configs

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNew(t *testing.T) {
	waypointRef := generics.GetFakeObj[*planv1.WaypointRef]()
	ecrRef := generics.GetFakeObj[*planv1.ECRRepositoryRef]()
	component := generics.GetFakeObj[*componentv1.Component]()

	// optional builders
	dockerCfg := generics.GetFakeObj[*buildv1.DockerConfig]()
	publicImg := generics.GetFakeObj[*PublicImageSource]()
	privateImg := generics.GetFakeObj[*PrivateImageSource]()

	tests := map[string]struct {
		optFns      func() []Option
		assertFn    func(*testing.T, *baseBuilder)
		errExpected error
	}{
		"happy path": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
				}
			},
			assertFn: func(t *testing.T, pln *baseBuilder) {
				assert.Equal(t, component, pln.Component)
				assert.Equal(t, waypointRef, pln.WaypointRef)
				assert.Equal(t, ecrRef, pln.EcrRef)
			},
		},
		"missing component": {
			optFns: func() []Option {
				return []Option{
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.Component"),
		},
		"missing waypoint ref": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithEcrRef(ecrRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.WaypointRef"),
		},
		"missing ecr ref": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithWaypointRef(waypointRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.EcrRef"),
		},
		"sets docker config": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
					WithDockerCfg(dockerCfg),
				}
			},
			assertFn: func(t *testing.T, pln *baseBuilder) {
				assert.True(t, proto.Equal(pln.DockerCfg, dockerCfg))
			},
		},
		"sets public image": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
					WithPublicImageSource(publicImg),
				}
			},
			assertFn: func(t *testing.T, pln *baseBuilder) {
				assert.Equal(t, pln.PublicImageSource, publicImg)
			},
		},
		"sets private image": {
			optFns: func() []Option {
				return []Option{
					WithComponent(component),
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
					WithPrivateImageSource(privateImg),
				}
			},
			assertFn: func(t *testing.T, pln *baseBuilder) {
				assert.Equal(t, pln.PrivateImageSource, privateImg)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			srv, err := newBaseBuilder(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, srv)
		})
	}
}
