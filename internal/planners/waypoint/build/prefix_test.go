package build

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/go-workflows-meta/prefix"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_planner_getPrefix(t *testing.T) {
	meta := generics.GetFakeObj[*planv1.Metadata]()
	orgMeta := generics.GetFakeObj[*planv1.OrgMetadata]()
	component := generics.GetFakeObj[*componentv1.Component]()

	pln, err := New(validator.New(), WithComponent(component), WithOrgMetadata(orgMeta), WithMetadata(meta))
	assert.NoError(t, err)

	assert.Equal(t, pln.getPrefix(), prefix.DeploymentPath(meta.OrgShortId,
		meta.AppShortId,
		component.Name,
		meta.DeploymentShortId))
}
