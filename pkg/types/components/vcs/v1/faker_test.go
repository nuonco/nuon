package vcsv1

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func Test_fakeVcsConfig(t *testing.T) {
	cfg, err := fakeVcsConfig(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	vcsCfg, ok := cfg.(*Config)
	assert.True(t, ok)
	err = vcsCfg.Validate()
	assert.NoError(t, err)
}

type testFakeObj struct {
	VcsConfig *Config `faker:"vcsConfig"`
}

func TestGetFakeObj(t *testing.T) {
	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)

	err = obj.VcsConfig.Validate()
	assert.NoError(t, err)
}
