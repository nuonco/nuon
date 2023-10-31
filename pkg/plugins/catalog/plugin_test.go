package catalog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPluginType(t *testing.T) {
	tests := map[string]struct {
		name        string
		typ         PluginType
		errExpected error
	}{
		"default": {
			name: "default",
			typ:  PluginTypeDefault,
		},
		"terraform": {
			name: "terraform",
			typ:  PluginTypeTerraform,
		},
		"exp": {
			name: "exp",
			typ:  PluginTypeExp,
		},
		"helm": {
			name: "helm",
			typ:  PluginTypeHelm,
		},
		"noop": {
			name: "noop",
			typ:  PluginTypeNoop,
		},
		"oci": {
			name: "oci",
			typ:  PluginTypeOci,
		},
		"oci-sync": {
			name: "oci-sync",
			typ:  PluginTypeOciSync,
		},
		"job": {
			name: "job",
			typ:  PluginTypeJob,
		},
		"invalid": {
			name:        "invalid",
			errExpected: fmt.Errorf("invalid"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pluginTyp, err := ToPluginType(test.name)

			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.typ, pluginTyp)
		})
	}
}

func TestPluginType_DevRepositoryName(t *testing.T) {
	tests := map[string]struct {
		typ      PluginType
		expected string
	}{
		"default": {
			typ:      PluginTypeDefault,
			expected: "waypoint-odr",
		},
		"terraform": {
			typ:      PluginTypeTerraform,
			expected: "dev-waypoint-plugin-terraform",
		},
		"exp": {
			typ:      PluginTypeExp,
			expected: "dev-waypoint-plugin-exp",
		},
		"noop": {
			typ:      PluginTypeNoop,
			expected: "dev-waypoint-plugin-noop",
		},
		"oci": {
			typ:      PluginTypeOci,
			expected: "dev-waypoint-plugin-oci",
		},
		"job": {
			typ:      PluginTypeJob,
			expected: "dev-waypoint-plugin-job",
		},
		"oci-sync": {
			typ:      PluginTypeOciSync,
			expected: "dev-waypoint-plugin-oci-sync",
		},
		"helm": {
			typ:      PluginTypeHelm,
			expected: "dev-waypoint-plugin-helm",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val := test.typ.DevRepositoryName()
			assert.Equal(t, test.expected, val)
		})
	}
}

func TestPluginType_RepositoryName(t *testing.T) {
	tests := map[string]struct {
		typ      PluginType
		expected string
	}{
		"default": {
			typ:      PluginTypeDefault,
			expected: "waypoint-odr",
		},
		"terraform": {
			typ:      PluginTypeTerraform,
			expected: "waypoint-plugin-terraform",
		},
		"exp": {
			typ:      PluginTypeExp,
			expected: "waypoint-plugin-exp",
		},
		"noop": {
			typ:      PluginTypeNoop,
			expected: "waypoint-plugin-noop",
		},
		"oci": {
			typ:      PluginTypeOci,
			expected: "waypoint-plugin-oci",
		},
		"oci-sync": {
			typ:      PluginTypeOciSync,
			expected: "waypoint-plugin-oci-sync",
		},
		"helm": {
			typ:      PluginTypeHelm,
			expected: "waypoint-plugin-helm",
		},
		"job": {
			typ:      PluginTypeJob,
			expected: "waypoint-plugin-job",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val := test.typ.RepositoryName()
			assert.Equal(t, test.expected, val)
		})
	}
}

func TestPluginType_ImageURL(t *testing.T) {
	tests := map[string]struct {
		typ      PluginType
		expected string
	}{
		"default": {
			typ:      PluginTypeDefault,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-odr",
		},
		"terraform": {
			typ:      PluginTypeTerraform,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-terraform",
		},
		"exp": {
			typ:      PluginTypeExp,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-exp",
		},
		"noop": {
			typ:      PluginTypeNoop,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-noop",
		},
		"helm": {
			typ:      PluginTypeHelm,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-helm",
		},
		"oci": {
			typ:      PluginTypeOci,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-oci",
		},
		"oci-sync": {
			typ:      PluginTypeOciSync,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-oci-sync",
		},
		"job": {
			typ:      PluginTypeJob,
			expected: "public.ecr.aws/p7e3r5y0/waypoint-plugin-job",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			val := test.typ.ImageURL()
			assert.Equal(t, test.expected, val)
		})
	}
}
