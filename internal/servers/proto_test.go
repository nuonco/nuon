package servers

import (
	"testing"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func Test_userModelToProto(t *testing.T) {
	org := generics.GetFakeObj[*models.Org]()
	pb, err := OrgModelToProto(org)

	assert.NoError(t, err)
	assert.Equal(t, org.ID.String(), pb.Id)
}
