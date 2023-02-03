package waypoint

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
	meta := generics.GetFakeObj[*planv1.Metadata]()
	orgMeta := generics.GetFakeObj[*planv1.OrgMetadata]()
	component := generics.GetFakeObj[*componentv1.Component]()

	tests := map[string]struct {
		optFns      func() []PlannerOption
		assertFn    func(*testing.T, *Planner)
		errExpected error
	}{
		"happy path": {
			optFns: func() []PlannerOption {
				return []PlannerOption{
					WithComponent(component),
					WithMetadata(meta),
					WithOrgMetadata(orgMeta),
				}
			},
			assertFn: func(t *testing.T, pln *Planner) {
				assert.Equal(t, component, pln.Component)
				assert.Equal(t, meta, pln.Metadata)
				assert.Equal(t, orgMeta, pln.OrgMetadata)
			},
		},
		"missing component": {
			optFns: func() []PlannerOption {
				return []PlannerOption{
					WithMetadata(meta),
					WithOrgMetadata(orgMeta),
				}
			},
			errExpected: fmt.Errorf("Planner.Component"),
		},
		"missing metadata": {
			optFns: func() []PlannerOption {
				return []PlannerOption{
					WithComponent(component),
					WithOrgMetadata(orgMeta),
				}
			},
			errExpected: fmt.Errorf("Planner.Metadata"),
		},
		"missing org metadata": {
			optFns: func() []PlannerOption {
				return []PlannerOption{
					WithComponent(component),
					WithMetadata(meta),
				}
			},
			errExpected: fmt.Errorf("Planner.OrgMetadata"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			srv, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, srv)
		})
	}
}
