package build

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
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

	prefix := pln.getPrefix()

	assert.Contains(t, prefix, fmt.Sprintf("org=%s", meta.OrgShortId))
	assert.Contains(t, prefix, fmt.Sprintf("app=%s", meta.AppShortId))
	assert.Contains(t, prefix, fmt.Sprintf("component=%s", component.Name))
	assert.Contains(t, prefix, fmt.Sprintf("deployment=%s", meta.DeploymentShortId))
}
