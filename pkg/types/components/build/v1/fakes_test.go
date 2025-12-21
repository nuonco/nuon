package buildv1

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func Test_fakeBuildConfig(t *testing.T) {
	cfg, err := fakeBuildConfig(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	buildCfg, ok := cfg.(*Config)
	assert.True(t, ok)
	err = buildCfg.Validate()
	assert.NoError(t, err)
}

type testFakeObj struct {
	BuildConfig *Config `faker:"buildConfig"`
}

func TestGetFakeObj(t *testing.T) {
	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)

	err = obj.BuildConfig.Validate()
	assert.NoError(t, err)
}
