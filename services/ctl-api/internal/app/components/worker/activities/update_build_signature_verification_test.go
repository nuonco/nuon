package activities

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/oci/signature"
)

func TestIsSignatureRejection(t *testing.T) {
	step := func(s string) pgtype.Hstore { return pgtype.Hstore{"step": &s} }

	require.True(t, isSignatureRejection(step(signature.VerifyStep)))
	require.False(t, isSignatureRejection(step("copy image")))
	require.False(t, isSignatureRejection(pgtype.Hstore{"step": nil}))
	require.False(t, isSignatureRejection(nil))
}
