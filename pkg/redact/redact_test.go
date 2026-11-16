package redact

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiagnosticSecrets(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"nuon token query": {
			input: "https://api.example.com/state?workspace_id=tfw123&token=tok_secret&operation=plan",
			want:  "https://api.example.com/state?workspace_id=tfw123&token=[REDACTED]&operation=plan",
		},
		"case insensitive multiple values": {
			input: "https://example.com?API_KEY=key-secret&Client-Secret=client-secret",
			want:  "https://example.com?API_KEY=[REDACTED]&Client-Secret=[REDACTED]",
		},
		"aws signed URL": {
			input: "https://s3.example.com/object?X-Amz-Credential=credential&X-Amz-Security-Token=session-token&X-Amz-Signature=signature",
			want:  "https://s3.example.com/object?X-Amz-Credential=[REDACTED]&X-Amz-Security-Token=[REDACTED]&X-Amz-Signature=[REDACTED]",
		},
		"azure SAS URL": {
			input: "https://account.blob.core.windows.net/container?sv=2025-01-05&sig=azure-signature&sp=r",
			want:  "https://account.blob.core.windows.net/container?sv=2025-01-05&sig=[REDACTED]&sp=r",
		},
		"gcp signed URL": {
			input: "https://storage.googleapis.com/object?X-Goog-Credential=credential&X-Goog-Signature=gcp-signature",
			want:  "https://storage.googleapis.com/object?X-Goog-Credential=[REDACTED]&X-Goog-Signature=[REDACTED]",
		},
		"URL userinfo": {
			input: "unable to clone https://user:github-token@github.com/acme/repo.git",
			want:  "unable to clone https://[REDACTED]@github.com/acme/repo.git",
		},
		"authorization and cookie headers": {
			input: "Authorization: Bearer bearer-token\nCookie: session=cookie-secret\nContent-Type: application/json",
			want:  "Authorization: [REDACTED]\nCookie: [REDACTED]\nContent-Type: application/json",
		},
		"JSON credential": {
			input: `request body: {"token":"json-secret","name":"safe"}`,
			want:  `request body: {"token":"[REDACTED]","name":"safe"}`,
		},
		"federation credentials": {
			input: `invalid_grant subject_token=acme-secret client_assertion=acme-secret {"client_assertion":"acme-secret"}`,
			want:  `invalid_grant subject_token=[REDACTED] client_assertion=[REDACTED] {"client_assertion":"[REDACTED]"}`,
		},
		"inline bearer": {
			input: "registry denied Bearer acme-secret; repository acme/api",
			want:  "registry denied Bearer [REDACTED]; repository acme/api",
		},
		"bare assignment": {
			input: "command failed with TOKEN=bare-secret",
			want:  "command failed with TOKEN=[REDACTED]",
		},
		"prefixed environment variables": {
			input: "AWS_SECRET_ACCESS_KEY=aws-secret GITHUB_TOKEN=github-secret DATABASE_PASSWORD=db-secret",
			want:  "AWS_SECRET_ACCESS_KEY=[REDACTED] GITHUB_TOKEN=[REDACTED] DATABASE_PASSWORD=[REDACTED]",
		},
		"prefixed JSON credential": {
			input: `request body: {"github_token":"json-secret","name":"safe"}`,
			want:  `request body: {"github_token":"[REDACTED]","name":"safe"}`,
		},
		"quoted HCL assignments": {
			input: `password = "hcl-secret" and api_token='single-quoted-secret'`,
			want:  `password = "[REDACTED]" and api_token='[REDACTED]'`,
		},
		"HTML encoded separator": {
			input: "https://example.com?first=safe&amp;token=html-secret&amp;last=safe",
			want:  "https://example.com?first=safe&amp;token=[REDACTED]&amp;last=safe",
		},
		"prose after token colon": {
			input: "unable to get authorization token: operation error ECR: GetAuthorizationToken: User: arn:aws:iam::111122223333:user/acme is not authorized",
			want:  "unable to get authorization token: operation error ECR: GetAuthorizationToken: User: arn:aws:iam::111122223333:user/acme is not authorized",
		},
		"prose after signature colon": {
			input: "signature: the request signature we calculated does not match",
			want:  "signature: the request signature we calculated does not match",
		},
		"prose after secret colon": {
			input: "secret: the client secret was not recognized by the registry",
			want:  "secret: the client secret was not recognized by the registry",
		},
		"colon without whitespace still redacted": {
			input: "https://example.com?token:colon-secret&first=safe",
			want:  "https://example.com?token:[REDACTED]&first=safe",
		},
		"benign URL": {
			input: "https://example.com?workspace_id=tfw123&operation=plan",
			want:  "https://example.com?workspace_id=tfw123&operation=plan",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DiagnosticSecrets(test.input); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestRedactionKeepsWrappedErrorCause(t *testing.T) {
	in := "unable to get ecr access info: unable to get authorization: unable to assume role: " +
		"arn:aws:iam::111122223333:role/example: no EC2 IMDS role found"

	got := DiagnosticSecrets(in)

	require.Equal(t, in, got, "a wrapped cause chain must survive redaction")
}

func TestRedactionStillMasksHeaders(t *testing.T) {
	got := DiagnosticSecrets("GET /v2/ HTTP/1.1\nAuthorization: Bearer supersecrettoken\nAccept: */*")

	require.NotContains(t, got, "supersecrettoken")
	require.True(t, strings.Contains(got, "Authorization: "))
}
