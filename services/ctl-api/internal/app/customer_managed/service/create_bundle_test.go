package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
)

func TestCanonicalBundleRunbooks(t *testing.T) {
	first := customermanaged.RunbookTemplate{ID: "first", Name: "First", Steps: []customermanaged.RunbookStep{{Kind: customermanaged.RunbookStepKindHealthGate}}}
	second := customermanaged.RunbookTemplate{ID: "second", Name: "Second", Steps: []customermanaged.RunbookStep{{Kind: customermanaged.RunbookStepKindAction, RefID: "action-a"}}}

	canonical, digest, err := canonicalBundleRunbooks([]customermanaged.RunbookTemplate{second, first})
	require.NoError(t, err)
	require.Equal(t, []customermanaged.RunbookTemplate{first, second}, canonical)

	_, reorderedDigest, err := canonicalBundleRunbooks([]customermanaged.RunbookTemplate{first, second})
	require.NoError(t, err)
	require.Equal(t, digest, reorderedDigest)

	second.Steps[0].RefID = "action-b"
	_, changedDigest, err := canonicalBundleRunbooks([]customermanaged.RunbookTemplate{first, second})
	require.NoError(t, err)
	require.NotEqual(t, digest, changedDigest)
}
