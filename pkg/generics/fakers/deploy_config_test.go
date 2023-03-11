package fakers

import (
	"reflect"
	"testing"

	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakeDeployConfig(t *testing.T) {
	cfg, err := fakeDeployConfig(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	deployCfg, ok := cfg.(*deployv1.Config)
	assert.True(t, ok)
	err = deployCfg.Validate()
	assert.NoError(t, err)
}
