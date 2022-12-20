package orgs

import (
	"testing"

	"github.com/powertoolsdev/api/internal/faker"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_userModelToProto(t *testing.T) {
	org := faker.GetFakeObj[*models.Org]()
	pb, err := orgModelToProto(org)

	assert.NoError(t, err)
	assert.Equal(t, org.ID.String(), pb.Id)
}
