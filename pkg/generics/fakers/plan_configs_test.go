package fakers

import (
	"reflect"
	"testing"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakePlanConfigs(t *testing.T) {
	cfg, err := fakePlanConfigs(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	cfgs, ok := cfg.([]*planv1.Config)
	assert.True(t, ok)

	for _, cfg := range cfgs {
		assert.NoError(t, cfg.Validate())
	}
}
