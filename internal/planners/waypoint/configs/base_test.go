package configs

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	waypointRef := generics.GetFakeObj[*planv1.WaypointRef]()
	ecrRef := generics.GetFakeObj[*planv1.ECRRepositoryRef]()
	component := generics.GetFakeObj[*componentv1.Component]()

	tests := map[string]struct {
		optFns      func() []baseBuilderOption
		assertFn    func(*testing.T, *baseBuilder)
		errExpected error
	}{
		"happy path": {
			optFns: func() []baseBuilderOption {
				return []baseBuilderOption{
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
			optFns: func() []baseBuilderOption {
				return []baseBuilderOption{
					WithWaypointRef(waypointRef),
					WithEcrRef(ecrRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.Component"),
		},
		"missing waypoint ref": {
			optFns: func() []baseBuilderOption {
				return []baseBuilderOption{
					WithComponent(component),
					WithEcrRef(ecrRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.WaypointRef"),
		},
		"missing ecr ref": {
			optFns: func() []baseBuilderOption {
				return []baseBuilderOption{
					WithComponent(component),
					WithWaypointRef(waypointRef),
				}
			},
			errExpected: fmt.Errorf("baseBuilder.EcrRef"),
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
