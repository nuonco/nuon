package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTemplatedValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"plain literal", "production", false},
		{"nuon reference", "{{ .nuon.install.id }}", true},
		{"nuon reference no spaces", "{{.nuon.components.api.outputs.version}}", true},
		{"mixed literal and template", "region-{{ .nuon.cloud_account.aws.region }}", true},
		{"braces without nuon", "{{ now }}", false},
		{"nuon without braces", ".nuon.install.id", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTemplatedValue(tt.value))
		})
	}
}

func TestSplitTemplated(t *testing.T) {
	l := Labels{
		"env":     "production",
		"region":  "{{ .nuon.cloud_account.aws.region }}",
		"version": "{{ .nuon.components.api.outputs.version }}",
		"team":    "platform",
	}

	static, templated := l.SplitTemplated()

	assert.Equal(t, Labels{"env": "production", "team": "platform"}, static)
	assert.Equal(t, Labels{
		"region":  "{{ .nuon.cloud_account.aws.region }}",
		"version": "{{ .nuon.components.api.outputs.version }}",
	}, templated)
}

func TestSplitTemplatedEmpty(t *testing.T) {
	static, templated := Labels(nil).SplitTemplated()
	assert.Empty(t, static)
	assert.Empty(t, templated)
}
