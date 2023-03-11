package fakers

import (
	"reflect"
	"testing"

	buildv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/build/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakeBuildConfig(t *testing.T) {
	cfg, err := fakeBuildConfig(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	buildCfg, ok := cfg.(*buildv1.Config)
	assert.True(t, ok)
	err = buildCfg.Validate()
	assert.NoError(t, err)
}
