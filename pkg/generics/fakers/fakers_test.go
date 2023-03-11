package fakers

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	buildv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/build/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/deploy/v1"
	"github.com/stretchr/testify/assert"
)

type testFakeObj struct {
	ShortID      string           `faker:"shortID"`
	BuildConfig  *buildv1.Config  `faker:"buildConfig"`
	DeployConfig *deployv1.Config `faker:"deployConfig"`
}

func TestGetFakeObj(t *testing.T) {
	Register()

	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)

	parsed, err := shortid.ToUUID(obj.ShortID)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed)

	err = obj.BuildConfig.Validate()
	assert.NoError(t, err)

	err = obj.DeployConfig.Validate()
	assert.NoError(t, err)
}
