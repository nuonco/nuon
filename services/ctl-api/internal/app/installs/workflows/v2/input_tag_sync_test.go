package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/generics"
)

func TestInputTagCandidates(t *testing.T) {
	got := inputTagCandidates(map[string]*string{
		"ctl_api_version": generics.ToPtr("0.19.1284"),
		"blank":           generics.ToPtr("  "),
		"sentence":        generics.ToPtr("not a tag"),
		"dup":             generics.ToPtr("0.19.1284"),
		"nil":             nil,
		"padded":          generics.ToPtr("  v1.2.3  "),
	})
	assert.Equal(t, []string{"0.19.1284", "v1.2.3"}, got)
	assert.Empty(t, inputTagCandidates(nil))
}
