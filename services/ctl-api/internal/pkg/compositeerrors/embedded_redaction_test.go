package compositeerrors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactionKeepsWrappedErrorCause(t *testing.T) {
	in := "unable to get ecr access info: unable to get authorization: unable to assume role: " +
		"arn:aws:iam::111122223333:role/example: no EC2 IMDS role found"

	got := RedactDiagnosticSecrets(in)

	require.Equal(t, in, got, "a wrapped cause chain must survive redaction")
}

func TestRedactionStillMasksHeaders(t *testing.T) {
	got := RedactDiagnosticSecrets("GET /v2/ HTTP/1.1\nAuthorization: Bearer supersecrettoken\nAccept: */*")

	require.NotContains(t, got, "supersecrettoken")
	require.True(t, strings.Contains(got, "Authorization: "))
}
