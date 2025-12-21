package deployv1

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func Test_fakeDeployConfig(t *testing.T) {
	cfg, err := fakeDeployConfig(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	deployCfg, ok := cfg.(*Config)
	assert.True(t, ok)
	err = deployCfg.Validate()
	assert.NoError(t, err)
}

type testFakeObj struct {
	DeployConfig *Config `faker:"deployConfig"`
}

func TestGetFakeObj(t *testing.T) {
	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)

	err = obj.DeployConfig.Validate()
	assert.NoError(t, err)
}
