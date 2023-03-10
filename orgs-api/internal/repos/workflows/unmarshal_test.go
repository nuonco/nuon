package workflows

import (
	"testing"

	"github.com/powertoolsdev/go-generics"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_parseResponse(t *testing.T) {
	fakeWorkflowResp := generics.GetFakeObj[*sharedv1.Response]()
	fakeWorkflowResp.Response = &sharedv1.ResponseRef{
		Response: generics.GetFakeObj[*sharedv1.ResponseRef_OrgSignup](),
	}

	byts, err := proto.Marshal(fakeWorkflowResp)
	assert.NoError(t, err)

	parsed, err := unmarshalResponse(byts)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(parsed, fakeWorkflowResp))
}
