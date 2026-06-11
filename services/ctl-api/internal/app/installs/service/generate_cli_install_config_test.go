package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
)

func TestBuildComponentOverridesFromInputs(t *testing.T) {
	ptr := func(s string) *string { return &s }

	values := map[string]*string{
		"replicas": ptr("3"), // real input, ignored
		config.HelmValuesOverrideInputName("clickhouse"): ptr("replicas: 5"),
		config.TFVarsOverrideInputName("vpc"):            ptr("cidr = \"10.0.0.0/16\""),
		config.TFVarsOverrideInputName("empty"):          ptr(""), // empty, omitted
	}

	got := config.Install{Components: buildComponentOverridesFromInputs(values)}.Components

	assert.Len(t, got, 2)
	assert.Equal(t, "replicas: 5", got["clickhouse"].HelmValues)
	assert.Equal(t, "cidr = \"10.0.0.0/16\"", got["vpc"].TFVars)
	_, hasEmpty := got["empty"]
	assert.False(t, hasEmpty, "empty override should be omitted")
}

func TestBuildComponentOverridesFromInputs_NoneReturnsNil(t *testing.T) {
	ptr := func(s string) *string { return &s }
	values := map[string]*string{"replicas": ptr("3")}
	assert.Nil(t, buildComponentOverridesFromInputs(values))
}
